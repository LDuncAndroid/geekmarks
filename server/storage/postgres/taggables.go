package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	hh "dmitryfrank.com/geekmarks/server/httphelper"
	"dmitryfrank.com/geekmarks/server/storage"
	"dmitryfrank.com/geekmarks/server/storage/postgres/internal/taghier"
	"github.com/juju/errors"
	_ "github.com/lib/pq"
)

func (s *StoragePostgres) CreateTaggable(tx *sql.Tx, tgbd *storage.TaggableData) (tgbID int, err error) {
	err = tx.QueryRow(
		"INSERT INTO taggables (owner_id, type) VALUES ($1, $2) RETURNING id",
		tgbd.OwnerID, string(tgbd.Type),
	).Scan(&tgbID)
	if err != nil {
		return 0, hh.MakeInternalServerError(errors.Annotatef(
			err, "adding new taggable (owner_id: %d, type: %s)", tgbd.OwnerID, tgbd.Type,
		))
	}

	return tgbID, nil
}

func (s *StoragePostgres) DeleteTaggable(tx *sql.Tx, taggableID int) error {
	_, err := tx.Exec(
		"DELETE FROM taggables WHERE id = $1", taggableID,
	)
	if err != nil {
		return hh.MakeInternalServerError(errors.Annotatef(
			err, "deleting taggable with id %d", taggableID,
		))
	}

	return nil
}

func (s *StoragePostgres) GetTaggedTaggableIDs(
	tx *sql.Tx, tagIDs []int, ownerID *int, ttypes []storage.TaggableType,
) (taggableIDs []int, err error) {
	args := []interface{}{}
	phNum := 1

	// Build query
	query := "SELECT id FROM taggables "

	for k, tagID := range tagIDs {
		query += fmt.Sprintf(
			"JOIN taggings t%d ON (t%d.taggable_id = taggables.id AND t%d.tag_id = $%d) ",
			k, k, k,
			phNum,
		)
		phNum++
		args = append(args, tagID)
	}

	query += "WHERE 1=1 "

	if ownerID != nil {
		query += fmt.Sprintf("AND owner_id = $%d ", phNum)
		phNum++
		args = append(args, *ownerID)
	}

	if len(ttypes) > 0 {
		qtmp := ""
		for i, ttype := range ttypes {
			if i > 0 {
				qtmp += "OR "
			}
			qtmp += fmt.Sprintf("type = $%d ", phNum)
			phNum++
			args = append(args, string(ttype))
		}
		query += "AND ( " + qtmp + " ) "
	}

	// Execute it
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, hh.MakeInternalServerError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var taggableID int
		err := rows.Scan(&taggableID)
		if err != nil {
			return nil, hh.MakeInternalServerError(err)
		}
		taggableIDs = append(taggableIDs, taggableID)
	}
	if err := rows.Close(); err != nil {
		return nil, errors.Annotatef(err, "closing rows")
	}

	return taggableIDs, nil
}

type tagBrief struct {
	ID       int    `json:"id"`
	ParentID int    `json:"parent_id"`
	Name     string `json:"name"`
}

type tagBriefMap map[string]tagBrief

func (tm tagBriefMap) GetParent(id int) (int, error) {
	t, ok := tm[strconv.Itoa(id)]
	if !ok {
		return 0, hh.MakeInternalServerError(errors.Errorf("no tag with id %d", id))
	}
	return t.ParentID, nil
}

func (tm tagBriefMap) GetPath(id int) (string, error) {
	t, ok := tm[strconv.Itoa(id)]
	if !ok {
		return "", hh.MakeInternalServerError(errors.Errorf("no tag with id %d", id))
	}

	ret := ""
	if t.ParentID != 0 {
		var err error
		ret, err = tm.GetPath(t.ParentID)
		if err != nil {
			return "", errors.Trace(err)
		}
		ret += "/"
	}

	return ret + t.Name, nil
}

func getTagsJsonFieldQuery(opts *storage.TagsFetchOpts, taggablesAlias string) (string, error) {
	switch opts.TagsFetchMode {
	case storage.TagsFetchModeNone:
		return "'{}'", nil
	case storage.TagsFetchModeLeafs, storage.TagsFetchModeAll:
		var nameArg, namesJoin string
		switch opts.TagNamesFetchMode {
		case storage.TagNamesFetchModeNone:
			nameArg = "''"
			namesJoin = ""
		case storage.TagNamesFetchModeShort, storage.TagNamesFetchModeFull:
			nameArg = "tn.name"
			namesJoin = `JOIN tag_names tn ON tags.id = tn.tag_id AND tn."primary" = 'true'`
		default:
			return "", errors.Errorf("wrong tag names fetch mode %q", opts.TagNamesFetchMode)
		}
		return fmt.Sprintf(`
       (
         SELECT JSON_OBJECT_AGG(
           tags.id,
           CAST(ROW(tags.id, tags.parent_id, %s) AS gm_tag_brief)
         )
         FROM taggings
         JOIN tags ON tags.id = taggings.tag_id
         %s
         WHERE taggings.taggable_id=%s.id
       )
		`, nameArg, namesJoin, taggablesAlias), nil
	default:
		return "", errors.Errorf("wrong tags fetch mode %q", opts.TagsFetchMode)
	}
}

func parseTagBrief(
	tagBriefData []byte, tagsFetchOpts *storage.TagsFetchOpts,
) (bmTags []storage.BookmarkTag, err error) {
	var tagBriefMap tagBriefMap
	if err := json.Unmarshal(tagBriefData, &tagBriefMap); err != nil {
		return nil, hh.MakeInternalServerError(err)
	}

	thier := taghier.New(tagBriefMap)
	for _, t := range tagBriefMap {
		thier.Add(t.ID)
	}

	var bkmTagIDs []int
	switch tagsFetchOpts.TagsFetchMode {
	case storage.TagsFetchModeLeafs:
		bkmTagIDs = thier.GetLeafs()

	case storage.TagsFetchModeAll:
		bkmTagIDs = thier.GetAll()
	}

	for _, tagID := range bkmTagIDs {
		tagBrief := tagBriefMap[strconv.Itoa(tagID)]

		var fullName string
		if tagsFetchOpts.TagNamesFetchMode == storage.TagNamesFetchModeFull {
			var err error
			fullName, err = tagBriefMap.GetPath(tagID)
			if err != nil {
				return nil, errors.Trace(err)
			}
		}

		bmTags = append(bmTags, storage.BookmarkTag{
			ID:       tagBrief.ID,
			ParentID: tagBrief.ParentID,
			Name:     tagBrief.Name,
			FullName: fullName,
		})
	}

	return bmTags, nil
}
