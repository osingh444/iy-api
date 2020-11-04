package models

import (
	"time"
	"fmt"
	"database/sql"

	"iybe/dbwrapper"
	"iybe/utils"
)

type ReviewItem struct {
	ReviewID      string
	VendorName    string
	Reviewer      string
	ReviewerID    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     time.Time
	ReviewText    string
	NumStars      uint8
}

type ReviewMediaItem struct {
	ReviewID      string
	VendorName    string
	S3URL1        string
	MediaID1      string
	S3URL2        string
	MediaID2      string
	S3URL3        string
	MediaID3      string
	CreatedAt     time.Time
	DeletedAt     time.Time
}

func (review *ReviewItem) CreateReview(claims *Claims) (map[string]interface{}, uint) {
	resp, ok := review.Validate()
	if !ok {
		return resp, 400
	}

	review.Reviewer = claims.DisplayName
	review.ReviewerID = claims.ID
	review.ReviewID = utils.GenerateUUID()
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	review.DeletedAt = time.Time{}

	txn, err := dbwrapper.GetDB().Begin()

	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	defer func() {
		_ = txn.Rollback()
  }()

	if _, err = txn.Exec(`INSERT INTO reviews (reviewID, reviewer, reviewerID, vendorName,
	                      numStars, createdAt, updatedAt, deletedAt, reviewText)
	                      VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, review.ReviewID, review.Reviewer,
												review.ReviewerID, review.VendorName, review.NumStars, review.CreatedAt,
												review.UpdatedAt, review.DeletedAt, review.ReviewText); err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	if _, err = txn.Exec("UPDATE vendors SET numReviews = numReviews + 1 WHERE vendorName = ?",
		                    review.VendorName); err != nil {
		return utils.Message("Server error"), 500
	}

	if err = txn.Commit(); err != nil {
		return utils.Message("Server error"), 500
	}

	response := utils.Message("Review created")
	response["rID"] = review.ReviewID
	return response, 201
}

func AddReviewMedia(urls []string, rid string, vendor string, mids []string) error {
	stmt, err := dbwrapper.GetDB().Prepare(`INSERT INTO media (reviewID, vendorName, s3URL1, mediaID1,
	                                        s3URL2, mediaID2, s3URL3, mediaID3, createdAt, deletedAt)
																					VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer stmt.Close()

	item := &ReviewMediaItem{
		ReviewID: rid,
		VendorName: vendor,
		S3URL1: "",
		MediaID1: "",
		S3URL2: "",
		MediaID2: "",
		S3URL3: "",
		MediaID3: "",
		CreatedAt: time.Now(),
		DeletedAt: time.Time{},
	}

	if len(urls) > 0 && len(mids) > 0 {
		item.S3URL1, item.MediaID1 = urls[0], mids[0]
		if len(urls) > 1 && len(mids) > 1 {
			item.S3URL2, item.MediaID2 = urls[1], mids[1]
			if len(urls) > 2 && len(mids) > 2 {
				item.S3URL3, item.MediaID3 = urls[2], mids[2]
			}
		}
	}

	if _, err = stmt.Exec(item.ReviewID, item.VendorName, item.S3URL1, item.MediaID1,
		                    item.S3URL2, item.MediaID2, item.S3URL3, item.MediaID3,
	                      item.CreatedAt, item.DeletedAt); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (review *ReviewItem) Validate() (map[string]interface{}, bool) {
	if review.NumStars < 1 || review.NumStars > 5 {
		return utils.Message("Invalid number of stars"), false
	}

	if review.ReviewText == "" {
		return utils.Message("Review cannot be empty"), false
	}

	if len(review.ReviewText) > 5000 {
		return utils.Message("Review text cannot be longer than 5000 characters"), false
	}

	if utils.ContainsBadWords(review.ReviewText) {
		return utils.Message("Review contains inappropriate language"), false
	}
	return nil, true
}

func DeleteReview(reviewID string, reviewerID string) (map[string]interface{}, uint){
	txn, err := dbwrapper.GetDB().Begin()

	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	defer func() {
		_ = txn.Rollback()
  }()

	row, err := txn.Query("SELECT * FROM reviews WHERE reviewID = ?", reviewID)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	reviews := make([]ReviewItem, 0)
	err = DBHandleReview(row, &reviews)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	row.Close()

	if len(reviews) == 0 {
		return utils.Message("Review not found"), 404
	}
	item := reviews[0]

	if item.ReviewerID != reviewerID {
		return utils.Message("Permission denied"), 403
	}

	if _, err = txn.Exec("DELETE FROM reviews WHERE reviewID = ?", reviewID); err != nil {
		return utils.Message("Server error"), 500
	}

	if err = DeleteReviewMedia(txn, reviewID); err != nil {
		return utils.Message("Server error"), 500
	}

	if _, err = txn.Exec("UPDATE vendors SET numReviews = numReviews - 1 WHERE vendorName = ?",
		                    item.VendorName); err != nil {
		return utils.Message("Server error"), 500
	}

	if err = txn.Commit(); err != nil {
		return utils.Message("Server error"), 500
	}

	return utils.Message("Review deleted successfully"), 200
}

func DeleteReviewMedia(txn *sql.Tx, reviewID string) error {
	row, err := txn.Query("SELECT * FROM media WHERE reviewID = ?", reviewID)

	if err != nil {
		return err
	}

	defer row.Close()

	var media ReviewMediaItem
	if err = DBHandleMediaItem(row, &media); err != nil {
		return err
	}

	if media.MediaID1 != "" {
		if err = utils.DeleteFromS3(media.MediaID1); err != nil {
			return err
		}
		if media.MediaID2 != "" {
			if err = utils.DeleteFromS3(media.MediaID2); err != nil {
				return err
			}
			if media.MediaID3 != "" {
				if err = utils.DeleteFromS3(media.MediaID3); err != nil {
					return err
				}
			}
		}
	}

	_, err = txn.Exec("DELETE FROM media WHERE reviewID = ?", reviewID)
	if err != nil {
		return err
	}

	return nil
}

func LoadReview(reviewerID string, vendor string) (ReviewItem, error) {
	var item ReviewItem
	row, err := dbwrapper.GetDB().Query("SELECT * FROM reviews WHERE reviewerID = ? AND vendorName = ?", reviewerID, vendor)

	if err != nil {
		return item, nil
	}

	defer row.Close()

	reviews := make([]ReviewItem, 0)
	err = DBHandleReview(row, &reviews)
	if err != nil {
		return item, nil
	}

	if len(reviews) == 0 {
		return item, nil
	}

	item = reviews[0]
	return item, nil
}

func LoadUserReviews(userID string) (map[string]interface{}, uint) {
	row, err := dbwrapper.GetDB().Query("SELECT * FROM reviews WHERE reviewerID = ?", userID)

	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	defer row.Close()

	reviews := make([]ReviewItem, 0)
	err = DBHandleReview(row, &reviews)
	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	response := utils.Message("Reviews retrieved")
	response["reviews"] = reviews

	return response, 200
}

func ReviewExists(rid string) bool {
	rows, err := dbwrapper.GetDB().Query("SELECT 1 FROM reviews WHERE reviewID=?", rid)

	if err != nil {
		return false
	}

	defer rows.Close()
	var exists int

	err = DBHandleInt(rows, &exists)

	if err != nil {
		return false
	}

	return exists == 1
}

func DBHandleReview(rows *sql.Rows, arr *[]ReviewItem) (error) {
	for rows.Next() {
		var review ReviewItem
		err := rows.Scan(&review.ReviewID, &review.Reviewer, &review.ReviewerID, &review.VendorName,
			               &review.NumStars, &review.CreatedAt, &review.UpdatedAt, &review.DeletedAt,
		                 &review.ReviewText)
		if err != nil {
			return err
		}
		*arr = append(*arr, review)
	}
	return nil
}

func DBHandleMediaItem(rows *sql.Rows, media *ReviewMediaItem) (error) {
	for rows.Next() {
		err := rows.Scan(&media.ReviewID, &media.VendorName, &media.S3URL1, &media.MediaID1,
			               &media.S3URL2, &media.MediaID2, &media.S3URL3, &media.MediaID3,
			               &media.CreatedAt, &media.DeletedAt)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
