package controllers

import (
	"iybe/models"
	"iybe/middleware"
	"iybe/utils"
	"encoding/json"
	"net/http"
	"fmt"
)

type QWrapper struct {
	Queue *utils.Queue
}

func (q *QWrapper) AddReport(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "user")
	if !ok {
		return
	}

	report := &models.Report{}
	err := json.NewDecoder(r.Body).Decode(report)

	if err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Invalid request"), 400)
		return
	}
	if q.Queue.Contains(report) {
		utils.Respond(w, utils.Message("report recieved"), 200)
		return
	}
	q.Queue.Add(report)
	utils.Respond(w, utils.Message("report added"), 201)
}

func ResolveReport(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "mod")
	if !ok {
		return
	}

	report := &models.Report{}
	err := json.NewDecoder(r.Body).Decode(report)

	if err != nil {
		utils.Respond(w, utils.Message("server error"), 500)
		return
	}

	resp, status := report.ResolveReport()
	utils.Respond(w, resp, status)
}

func (q *QWrapper) LoadReports(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "mod")
	if !ok {
		return
	}

	reports := make([]models.ReviewItem, 0)
	qreps := q.Queue.RemoveUpTo(10)

	for _, rep := range qreps {
		item := rep.(*models.Report)
		rev, err := models.LoadReview(item.ReviewerID, item.VendorName)
		if err != nil {
			continue
		}
		reports = append(reports, rev)
	}
	resp := utils.Message("Reviews retrieved")
	resp["reports"] = reports
	resp["qports"] = qreps

	utils.Respond(w, resp, 200)
}

func (q *QWrapper) SerializeReports(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.AuthenticateToken(w, r, "mod")
	if !ok {
		return
	}

	err := q.Queue.Serialize("q.txt")
	if err != nil {
		utils.Respond(w, utils.Message("Serialization failed"), 500)
		return
	}
	
	utils.Respond(w, utils.Message("Serialized queue"), 200)
}
