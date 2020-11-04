package controllers

import (
	"iybe/models"
	"iybe/utils"

	"strconv"
	"net/http"
	"encoding/json"
	"strings"
	"fmt"
)

func AddVendor(w http.ResponseWriter, r *http.Request) {
	vendor := &models.AddVendorItem{}
	err := json.NewDecoder(r.Body).Decode(vendor)
	if err != nil {
		fmt.Println(err)
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	vendor.IGName = strings.ToLower(vendor.IGName)

	if !vendor.IsValidVendorName() {
		utils.Respond(w, utils.Message("Invalid vendor name supplied"), 400)
		return
	}

	if models.VendorExists(vendor.IGName) {
		utils.Respond(w, utils.Message("Valid vendor has already been added"), 200)
		return
	} else if vendor.Validate() {
		resp, status := vendor.CreateVendor()
		utils.Respond(w, resp, status)
		return
	} else {
		utils.Respond(w, utils.Message("Invalid vendor"), 404)
		return
	}

	utils.Respond(w, utils.Message("Something went wrong"), 404)
}

func LoadReviews(w http.ResponseWriter, r *http.Request) {
	vvendor := r.URL.Query()["v"]
	oofset := r.URL.Query()["offset"]
	if len(vvendor) != 1 || len(oofset) != 1 {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}

	vendor := vvendor[0]
	offset := oofset[0]
	if vendor == "" || offset == ""{
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	i, err := strconv.Atoi(offset)
	if err != nil {
		utils.Respond(w, utils.Message("Invalid Request"), 400)
		return
	}
	
	resp, status := models.GetVendorReviews(vendor, i)
	utils.Respond(w, resp, status)
}
