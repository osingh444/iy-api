package controllers

import (
	"iybe/models"
	"iybe/middleware"
	"iybe/utils"

	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"bytes"
	"path"
)

var maxPhotoSize int64 = 500000000

func AddReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.AuthenticateToken(w, r, "user")
	if !ok {
		return
	}

	review := &models.ReviewItem{}
	err := json.NewDecoder(r.Body).Decode(review)

	if err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Invalid request"), 400)
		return
	}

	if !models.VendorExists(review.VendorName) {
		utils.Respond(w, utils.Message("Review added for vendor that does not exist"), 400)
		return
	}

	resp, status := review.CreateReview(claims)

	utils.Respond(w, resp, status)
}

func AddReviewPhotos(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "user")
	if !ok {
		return
	}

	rID := r.URL.Query()["rid"]
	vvendor := r.URL.Query()["v"]
	if len(rID) != 1 || len(vvendor) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	reviewID := rID[0]
	vendor := vvendor[0]

	if reviewID == "" || vendor == "" {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	if !models.ReviewExists(reviewID) {
		utils.Respond(w, utils.Message("Media added for nonexistent review"), 400)
		return
	}

	err := r.ParseMultipartForm(10000000)
  if err != nil {
		fmt.Println(err)
    utils.Respond(w, utils.Message("Invalid request"), 400)
		return
  }

	files := r.MultipartForm.File["files[]"]
	urls := make([]string, 0)
	mids := make([]string, 0)
	for i, _ := range files {
		if(i > 2) {
			break
		}

		fileSize := files[i].Size
		if !utils.IsAllowedFileSize(fileSize, maxPhotoSize) {
			utils.Respond(w, utils.Message("File exceeds allowed size"), 400)
			return
		}

		file, err := files[i].Open()
    defer file.Close()

    if err != nil {
			fmt.Println(err)
      utils.Respond(w, utils.Message("Invalid photos"), 400)
      return
    }

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
      utils.Respond(w, utils.Message("Server error"), 500)
      return
    }

		fileType := http.DetectContentType(fileBytes)
		if ok := utils.IsValidPhotoFormat(fileType); !ok {
			utils.Respond(w, utils.Message("Invalid file type"), 400)
			return
		}

		finalName := fmt.Sprintf("%s%s", utils.GenerateUUID(), path.Ext(files[i].Filename))
		fileReader := bytes.NewReader(fileBytes)
		if utils.PutToS3(finalName, fileType, fileReader, fileSize) != nil {
			utils.Respond(w, utils.Message("Server error"), 500)
			return
		}

		urls = append(urls, utils.GenerateS3URL(finalName))
		mids = append(mids, finalName)
	}

	if err = models.AddReviewMedia(urls, reviewID, vendor, mids); err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Server error"), 500)
		return
	}

	utils.Respond(w, utils.Message("Media added"), 201)
}

func DeleteReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.AuthenticateToken(w, r, "user")
	if !ok {
		return
	}

	rID := r.URL.Query()["rid"]
	if len(rID) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	reviewID := rID[0]

	if reviewID == "" {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	resp, status := models.DeleteReview(reviewID, claims.ID)

	utils.Respond(w, resp, status)
}

func LoadUserReviews(w http.ResponseWriter, r *http.Request) {
	uuser := r.URL.Query()["u"]
	if len(uuser) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	user := uuser[0]

	if user == "" {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	resp, status := models.LoadUserReviews(user)
	utils.Respond(w, resp, status)
}
