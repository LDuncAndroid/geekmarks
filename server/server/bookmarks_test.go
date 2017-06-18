// Copyright 2017 Dmitry Frank <mail@dmitryfrank.com>
// Licensed under the BSD, see LICENSE file for details.

// +build all_tests integration_tests

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"dmitryfrank.com/geekmarks/server/cptr"
	"dmitryfrank.com/geekmarks/server/storage"
	"github.com/juju/errors"
)

// Test bookmarks {{{
func TestBookmarks(t *testing.T) {
	runWithRealDB(t, func(si storage.Storage, be testBackend) error {
		var err error

		err = runPerUserTest(si, be, "test1", "1@1.1", "test2", "2@1.1", perUserTestBookmarks)
		if err != nil {
			return errors.Trace(err)
		}

		return nil
	})
}

func perUserTestBookmarks(
	si storage.Storage, be testBackend, u1, u2 *perUserData,
) error {
	var err error

	tagIDs, err := makeTestTagsHierarchy(be, u1.id)
	if err != nil {
		return errors.Trace(err)
	}

	// add bkm1 tagged with tag1/tag3 and tag8
	bkm1ID, err := addBookmark(be, u1.id, &bkmData{
		URL:     "url_1",
		Title:   "title_1",
		Comment: "comment_1",
		TagIDs: []int{
			tagIDs.tag3ID,
			tagIDs.tag8ID,
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	// add bkm2 tagged with tag1
	bkm2ID, err := addBookmark(be, u1.id, &bkmData{
		URL:     "url_2",
		Title:   "title_2",
		Comment: "comment_2",
		TagIDs: []int{
			tagIDs.tag1ID,
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	// add bkm3, untagged
	bkm3ID, err := addBookmark(be, u1.id, &bkmData{
		URL:     "url_3",
		Title:   "title_3",
		Comment: "comment_3",
		TagIDs:  []int{},
	})
	if err != nil {
		return errors.Trace(err)
	}

	err = checkBkmGetByID(be, u1.id, bkm1ID, &bkmData{
		ID:      bkm1ID,
		URL:     "url_1",
		Title:   "title_1",
		Comment: "comment_1",
		Tags: []bkmTagData{
			bkmTagData{
				Items: []bkmTagDataItem{
					bkmTagDataItem{
						ID:   tagIDs.tag1ID,
						Name: "tag1",
					},
					bkmTagDataItem{
						ID:   tagIDs.tag3ID,
						Name: "tag3_alias",
					},
				},
			},
			bkmTagData{
				Items: []bkmTagDataItem{
					bkmTagDataItem{
						ID:   tagIDs.tag7ID,
						Name: "tag7",
					},
					bkmTagDataItem{
						ID:   tagIDs.tag8ID,
						Name: "tag8",
					},
				},
			},
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag3: should return bkm1
	bkmRespData, err := checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag3ID}}, []int{bkm1ID},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// check contents as well
	{
		errStr := ""
		if got, want := bkmRespData[0].URL, "url_1"; got != want {
			errStr += fmt.Sprintf("bookmark url: got %q, want %q\n", got, want)
		}
		if got, want := bkmRespData[0].Title, "title_1"; got != want {
			errStr += fmt.Sprintf("bookmark title: got %q, want %q\n", got, want)
		}
		if got, want := bkmRespData[0].Comment, "comment_1"; got != want {
			errStr += fmt.Sprintf("bookmark comment: got %q, want %q\n", got, want)
		}
		if errStr != "" {
			return errors.New(errStr)
		}
	}

	if err := checkBkmTags(&bkmRespData[0], []bkmTagData{
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag1ID,
					Name: "tag1",
				},
				bkmTagDataItem{
					ID:   tagIDs.tag3ID,
					Name: "tag3_alias",
				},
			},
		},
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag7ID,
					Name: "tag7",
				},
				bkmTagDataItem{
					ID:   tagIDs.tag8ID,
					Name: "tag8",
				},
			},
		},
	},
	); err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag1: should return bkm1, bkm2
	bkmRespData, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag1ID}}, []int{bkm1ID, bkm2ID},
	)
	if err != nil {
		return errors.Trace(err)
	}

	if err := checkBkmTags(&bkmRespData[0], []bkmTagData{
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag1ID,
					Name: "tag1",
				},
				bkmTagDataItem{
					ID:   tagIDs.tag3ID,
					Name: "tag3_alias",
				},
			},
		},
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag7ID,
					Name: "tag7",
				},
				bkmTagDataItem{
					ID:   tagIDs.tag8ID,
					Name: "tag8",
				},
			},
		},
	}); err != nil {
		return errors.Trace(err)
	}

	if err := checkBkmTags(&bkmRespData[1], []bkmTagData{
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag1ID,
					Name: "tag1",
				},
			},
		},
	}); err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag1, tag3: should return bkm1
	_, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag1ID, tagIDs.tag3ID}}, []int{bkm1ID},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag1, tag3, tag2: should return nothing
	_, err = checkBkmGet(
		be, u1.id,
		&bkmGetArg{tagIDs: []int{tagIDs.tag1ID, tagIDs.tag3ID, tagIDs.tag2ID}},
		[]int{},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// add bkm2 tagged with tag1
	// update bkm2: now, it's tagged with tag7/tag8
	err = updateBookmark(be, u1.id, &bkmData{
		ID:      bkm2ID,
		URL:     "url_2_upd",
		Title:   "title_2_upd",
		Comment: "comment_2_upd",
		TagIDs: []int{
			tagIDs.tag8ID,
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag7: should return bkm1, bkm2
	bkmRespData, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag7ID}}, []int{bkm1ID, bkm2ID},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// check contents as well
	{
		errStr := ""
		if got, want := bkmRespData[1].URL, "url_2_upd"; got != want {
			errStr += fmt.Sprintf("bookmark url: got %q, want %q\n", got, want)
		}
		if got, want := bkmRespData[1].Title, "title_2_upd"; got != want {
			errStr += fmt.Sprintf("bookmark title: got %q, want %q\n", got, want)
		}
		if got, want := bkmRespData[1].Comment, "comment_2_upd"; got != want {
			errStr += fmt.Sprintf("bookmark comment: got %q, want %q\n", got, want)
		}
		if errStr != "" {
			return errors.New(errStr)
		}
	}

	if err := checkBkmTags(&bkmRespData[1], []bkmTagData{
		bkmTagData{
			Items: []bkmTagDataItem{
				bkmTagDataItem{
					ID:   tagIDs.tag7ID,
					Name: "tag7",
				},
				bkmTagDataItem{
					ID:   tagIDs.tag8ID,
					Name: "tag8",
				},
			},
		},
	}); err != nil {
		return errors.Trace(err)
	}

	// get untagged: should return bkm3
	_, err = checkBkmGet(be, u1.id, &bkmGetArg{tagIDs: []int{}}, []int{bkm3ID})
	if err != nil {
		return errors.Trace(err)
	}

	// get bookmark by url
	bkmRespData, err = checkBkmGet(
		be, u1.id, &bkmGetArg{url: cptr.String("url_2_upd")}, []int{bkm2ID},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// get bookmark by non-existing url
	bkmRespData, err = checkBkmGet(
		be, u1.id, &bkmGetArg{url: cptr.String("non-existing-url")}, []int{},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// try to add bookmark with the existing URL, should fail
	bkm100ID, err := addBookmark(be, u1.id, &bkmData{
		URL:     "url_1",
		Title:   "title_100",
		Comment: "comment_100",
		TagIDs:  []int{},
	})
	if err == nil || bkm100ID != 0 {
		return errors.Errorf("adding the bookmark with existing URL %q should fail", "url_1")
	}

	// try to add bookmark with the existing URL, should fail
	bkm100ID, err = addBookmark(be, u1.id, &bkmData{
		URL:     "url_1",
		Title:   "title_100",
		Comment: "comment_100",
		TagIDs:  []int{},
	})
	if err == nil || bkm100ID != 0 {
		return errors.Errorf("adding the bookmark with existing URL %q should fail", "url_1")
	}

	// try to update URL with the same data
	err = updateBookmark(be, u1.id, &bkmData{
		ID:      bkm2ID,
		URL:     "url_2_upd",
		Title:   "title_2_upd",
		Comment: "comment_2_upd",
		TagIDs: []int{
			tagIDs.tag8ID,
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	// try to update URL of the bookmark to the existing one (should fail)
	err = updateBookmark(be, u1.id, &bkmData{
		ID:      bkm2ID,
		URL:     "url_1",
		Title:   "title_2_upd",
		Comment: "comment_2_upd",
		TagIDs: []int{
			tagIDs.tag8ID,
		},
	})
	if err == nil {
		return errors.Errorf("updating the bookmark's URL to the existing value %q should fail", "url_1")
	}

	fmt.Println(tagIDs.tag1ID, bkm1ID, bkm2ID, bkm3ID, bkm100ID)

	return nil
}

// }}}

// Test deletion of bookmarks {{{
func TestDeleteBookmarks(t *testing.T) {
	runWithRealDB(t, func(si storage.Storage, be testBackend) error {
		var err error

		err = runPerUserTest(si, be, "test1", "1@1.1", "test2", "2@1.1", perUserTestDeleteBookmarks)
		if err != nil {
			return errors.Trace(err)
		}

		return nil
	})
}

func perUserTestDeleteBookmarks(
	si storage.Storage, be testBackend, u1, u2 *perUserData,
) error {
	var err error

	tagIDs, err := makeTestTagsHierarchy(be, u1.id)
	if err != nil {
		return errors.Trace(err)
	}

	bkmIDs, err := makeTestBookmarks(be, u1.id, tagIDs)
	if err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag3
	_, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag3ID}}, []int{
			bkmIDs.bkm3ID,
			bkmIDs.bkm4ID,
			bkmIDs.bkm5ID,
			bkmIDs.bkm6ID,
			bkmIDs.bkm2_5ID,
			bkmIDs.bkm4_5ID,
		},
	)
	if err != nil {
		return errors.Trace(err)
	}

	if err := deleteBookmark(be, u1.id, bkmIDs.bkm2_5ID); err != nil {
		return errors.Trace(err)
	}

	// get tagged with tag3 (there should be no bookmark which we've just deleted)
	_, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag3ID}}, []int{
			bkmIDs.bkm3ID,
			bkmIDs.bkm4ID,
			bkmIDs.bkm5ID,
			bkmIDs.bkm6ID,
			bkmIDs.bkm4_5ID,
		},
	)
	if err != nil {
		return errors.Trace(err)
	}

	// test that deleting a bookmark of another user is forbidden
	{
		resp, err := be.DoReq(
			"DELETE", fmt.Sprintf("/api/users/%d/bookmarks/%d", u1.id, bkmIDs.bkm6ID), u2.token,
			nil, false,
		)
		if err != nil {
			return errors.Trace(err)
		}
		if err := expectErrorResp(resp, http.StatusForbidden, "forbidden"); err != nil {
			return errors.Trace(err)
		}
	}

	// get tagged with tag3 (should not change)
	_, err = checkBkmGet(
		be, u1.id, &bkmGetArg{tagIDs: []int{tagIDs.tag3ID}}, []int{
			bkmIDs.bkm3ID,
			bkmIDs.bkm4ID,
			bkmIDs.bkm5ID,
			bkmIDs.bkm6ID,
			bkmIDs.bkm4_5ID,
		},
	)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// }}}

type bkmData struct {
	ID        int          `json:"id"`
	URL       string       `json:"url"`
	Title     string       `json:"title,omitempty"`
	Comment   string       `json:"comment,omitempty"`
	UpdatedAt uint64       `json:"updatedAt"`
	TagIDs    []int        `json:"tagIDs"`
	Tags      []bkmTagData `json:"tags,omitempty"`
}

// bkmTagsByID implements sorting by the last tag item ID
type bkmTagsByID []bkmTagData

type bkmTagData struct {
	Items []bkmTagDataItem `json:"items"`
}

type bkmTagDataItem struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

func (s bkmTagsByID) Len() int {
	return len(s)
}
func (s bkmTagsByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bkmTagsByID) Less(i, j int) bool {
	ilen := len(s[i].Items)
	jlen := len(s[j].Items)
	return s[i].Items[ilen-1].ID < s[j].Items[jlen-1].ID
}

type bkms []bkmData
type bkmsByID bkms

func (s bkmsByID) Len() int {
	return len(s)
}
func (s bkmsByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s bkmsByID) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func addBookmark(be testBackend, userID int, data *bkmData) (bkmID int, err error) {
	tagIDs := A{}
	for _, id := range data.TagIDs {
		tagIDs = append(tagIDs, id)
	}
	resp, err := be.DoUserReq("POST", "/bookmarks", userID, H{
		"url":     data.URL,
		"title":   data.Title,
		"comment": data.Comment,
		"tagIDs":  tagIDs,
	}, true)
	if err != nil {
		return 0, errors.Trace(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Trace(err)
	}

	v := map[string]int{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return v["bookmarkID"], nil
}

func updateBookmark(be testBackend, userID int, data *bkmData) (err error) {
	tagIDs := A{}
	for _, id := range data.TagIDs {
		tagIDs = append(tagIDs, id)
	}
	resp, err := be.DoUserReq("PUT", fmt.Sprintf("/bookmarks/%d", data.ID), userID, H{
		"url":     data.URL,
		"title":   data.Title,
		"comment": data.Comment,
		"tagIDs":  tagIDs,
	}, true)
	if err != nil {
		return errors.Trace(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	v := map[string]int{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func deleteBookmark(be testBackend, userID, bkmID int) (err error) {
	_, err = be.DoUserReq(
		"DELETE", fmt.Sprintf("/bookmarks/%d", bkmID),
		userID, nil, true,
	)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

type bkmGetArg struct {
	tagIDs []int
	url    *string
}

func checkBkmGet(
	be testBackend, userID int, args *bkmGetArg, expectedBkmIDs []int,
) ([]bkmData, error) {

	//qsParts := []string{}
	//for _, tagID := range tagIDs {
	//qsParts = append(qsParts, fmt.Sprintf("tag_id=%d", tagID))
	//}

	qsVals := url.Values{}
	if args.url != nil {
		qsVals.Add("url", *args.url)
	} else {
		for _, tagID := range args.tagIDs {
			qsVals.Add("tag_id", strconv.Itoa(tagID))
		}
	}

	resp, err := be.DoUserReq(
		"GET", "/bookmarks?"+qsVals.Encode(), userID, nil, true,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	v := bkms{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Printf("body: %q\n", body)
		return nil, errors.Trace(err)
	}

	bkmIDs := []int{}
	for _, b := range v {
		bkmIDs = append(bkmIDs, b.ID)
	}

	sort.Ints(bkmIDs)
	sort.Ints(expectedBkmIDs)

	if !reflect.DeepEqual(bkmIDs, expectedBkmIDs) {
		return nil, errors.Errorf("bookmarks mismatch: expected %v, got %v", expectedBkmIDs, bkmIDs)
	}

	sort.Sort(bkmsByID(v))

	return []bkmData(v), nil
}

func checkBkmGetByID(be testBackend, userID int, bkmID int, expectedBkm *bkmData) error {
	resp, err := be.DoUserReq(
		"GET", fmt.Sprintf("/bookmarks/%d", bkmID), userID, nil, true,
	)
	if err != nil {
		return errors.Trace(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	v := bkmData{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Printf("body: %q\n", body)
		return errors.Trace(err)
	}

	// don't compare UpdatedAt
	v.UpdatedAt = 0

	sort.Sort(bkmTagsByID(v.Tags))
	sort.Sort(bkmTagsByID(expectedBkm.Tags))

	if !reflect.DeepEqual(&v, expectedBkm) {
		return errors.Errorf("bookmark mismatches: expected %v, got %v", expectedBkm, v)
	}

	return nil
}

func checkBkmTags(bkm *bkmData, expectedTags []bkmTagData) error {
	sort.Sort(bkmTagsByID(expectedTags))
	sort.Sort(bkmTagsByID(bkm.Tags))

	if !reflect.DeepEqual(expectedTags, bkm.Tags) {
		return errors.Errorf("bookmark tags mismatch: expected %v, got %v", expectedTags, bkm.Tags)
	}

	return nil
}
