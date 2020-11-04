package models

import (
	"iybe/utils"
	"iybe/dbwrapper"

	"database/sql"
	"bytes"
	"regexp"
	"time"
	"strings"
	"fmt"
)

type AddVendorItem struct {
	IGName string
	Page   string
}

type VendorItem struct {
	VendorName string
	Reviews    int
	Num1Star   int
	Num2Star   int
	Num3Star   int
	Num4Star   int
	Num5Star   int

}

type VendorReview struct {
	Review    ReviewItem
	Media     ReviewMediaItem
}

func (vendor *AddVendorItem) IsValidVendorName() bool {
	if len(vendor.IGName) > 30 || len(vendor.IGName) == 0 {
		return false
	}

	//byte(".") is 46
	if vendor.IGName[0] == 46 || vendor.IGName[len(vendor.IGName) - 1] == 46 {
		return false
	}

	isOk := regexp.MustCompile(`^[a-z0-9_.]$`).MatchString

	for i := 0; i < len(vendor.IGName); i++ {
		if !isOk(string(vendor.IGName[i])) {
			return false
		}
		if vendor.IGName[i] == 46 && vendor.IGName[i + 1] == 46 {
			return false
		}
	}
	return true
}

func VendorExists(vendor string) (bool) {
	rows, err := dbwrapper.GetDB().Query("SELECT 1 FROM vendors WHERE vendorName=?", vendor)

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

func (vendor *AddVendorItem) Validate() bool {
	return vendor.ValidateTitle()
}

func (vendor *AddVendorItem) ValidateTitle() bool {
	parts := strings.Split(vendor.Page, "<title>")
	if len(parts) != 2 {
		return false
	}
	parts2 := strings.Split(parts[1], "</title>")
	if len(parts) != 2 {
		return false
	}
	fmt.Println(parts2[1])
	return strings.Contains(parts2[0], "(@" + vendor.IGName + ")")
}

func (vendor *AddVendorItem) CreateVendor() (map[string]interface{}, uint) {
	vendorItem := &VendorItem{
		VendorName: vendor.IGName,
		Reviews: 0,
		Num1Star: 0,
		Num2Star: 0,
		Num3Star: 0,
		Num4Star: 0,
		Num5Star: 0,
	}

	stmt, err := dbwrapper.GetDB().Prepare(`INSERT INTO vendors (vendorName, createdAt, updatedAt, deletedAt, numReviews)
	                                     	VALUES( ?, ?, ?, ?, ?)`)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	defer stmt.Close()

	if _, err = stmt.Exec(&vendorItem.VendorName, time.Now(), time.Now(), time.Time{}, &vendorItem.Reviews); err != nil {
		return utils.Message("Server error"), 500
	}

	return utils.Message("Vendor created"), 201
}

//adds to the number of star stats counter
func UpdateVendorStats(vendor string, howManyStars uint8, operation string) (map[string]interface{}, uint) {
	return nil, 200
}

func GetVendorReviews(vendor string, offset int) (map[string]interface{}, uint) {
	if !VendorExists(vendor) {
		return utils.Message("vendor does not exist"), 404
	}

	row, err := dbwrapper.GetDB().Query(`SELECT * FROM reviews WHERE vendorName = ?
		                                   ORDER BY createdAt DESC LIMIT ? OFFSET ?`, vendor, 5, offset)

	if err != nil {
		fmt.Println(err)
		return utils.Message("Server error"), 500
	}

	reviews := make([]VendorReview, 0)
	revIDs := make([]interface{}, 0)
	revItems := make([]ReviewItem, 0)
	mediaItems := make([]ReviewMediaItem, 0)

	if err = DBHandleReview(row, &revItems); err != nil {
		return utils.Message("Server error"), 500
	}

	row.Close()

	var qbytes bytes.Buffer
	qbytes.WriteString("SELECT * FROM media WHERE reviewID = ?")

	for index, rev := range revItems {
		if index > 0 {
			qbytes.WriteString(" OR reviewID = ?")
		}

		var review VendorReview
		review.Review = rev

		reviews = append(reviews, review)
		revIDs = append(revIDs, rev.ReviewID)
	}

	res, err := dbwrapper.GetDB().Query(qbytes.String(), revIDs...)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	if err = DBHandleMediaItems(res, &mediaItems); err != nil {
		return utils.Message("Server error"), 500
	}

	res.Close()

	for index, review := range reviews {
		for  _, media := range mediaItems {
			if media.ReviewID == review.Review.ReviewID {
				reviews[index].Media = media
				break
			}
		}
	}

	row, err = dbwrapper.GetDB().Query("SELECT numReviews FROM vendors WHERE vendorName = ?", vendor)
	if err != nil {
		return utils.Message("Server error"), 500
	}

	var totalRevs int
	if err = DBHandleInt(row, &totalRevs); err != nil {
		return utils.Message("Server error"), 500
	}

	row.Close()

	response := utils.Message("Reviews retrieved")
	response["reviews"] = reviews
	response["count"] = totalRevs

	return response, 200
}

func DBHandleMediaItems(rows *sql.Rows, arr *[]ReviewMediaItem) (error) {
	for rows.Next() {
		var media ReviewMediaItem
		err := rows.Scan(&media.ReviewID, &media.VendorName, &media.S3URL1, &media.MediaID1,
			               &media.S3URL2, &media.MediaID2, &media.S3URL3, &media.MediaID3,
			               &media.CreatedAt, &media.DeletedAt)

		if err != nil {
			fmt.Println(err)
			return err
		}
		*arr = append(*arr, media)
	}
	return nil
}
