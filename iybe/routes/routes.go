package routes

import (
	"iybe/controllers"

	"github.com/gorilla/mux"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"

	"time"
)

func Routes(q *controllers.QWrapper) *mux.Router {
	router := mux.NewRouter()

	//api version
	s := router.PathPrefix("/v1").Subrouter()
	lmt := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})

	//account routes
	s.Handle("/register", tollbooth.LimitFuncHandler(lmt, controllers.CreateUser)).Methods("POST")
	s.Handle("/login", tollbooth.LimitFuncHandler(lmt, controllers.Authenticate)).Methods("POST")
	s.Handle("/confirm", tollbooth.LimitFuncHandler(lmt, controllers.ConfirmEmail)).Methods("GET")
	s.Handle("/reqreset", tollbooth.LimitFuncHandler(lmt, controllers.RequestPasswordReset)).Methods("GET")
	s.Handle("/reset", tollbooth.LimitFuncHandler(lmt, controllers.UpdateUserPassword)).Methods("POST")

	//review routes
	s.Handle("/createreview", tollbooth.LimitFuncHandler(lmt, controllers.AddReview)).Methods("POST")
	s.Handle("/addreviewphoto", tollbooth.LimitFuncHandler(lmt, controllers.AddReviewPhotos)).Methods("POST")
	s.Handle("/deletereview", tollbooth.LimitFuncHandler(lmt, controllers.DeleteReview)).Methods("DELETE")

	//vendor routes
	s.Handle("/addvendor", tollbooth.LimitFuncHandler(lmt, controllers.AddVendor)).Methods("POST")
	s.Handle("/vendor", tollbooth.LimitFuncHandler(lmt, controllers.LoadReviews)).Methods("GET")

	//user routes
	s.Handle("/userreviews", tollbooth.LimitFuncHandler(lmt, controllers.LoadUserReviews)).Methods("GET")

	//report routes
	s.Handle("/addreport", tollbooth.LimitFuncHandler(lmt, q.AddReport)).Methods("POST")
	s.Handle("/loadreports", tollbooth.LimitFuncHandler(lmt, q.LoadReports)).Methods("GET")
	s.Handle("/resolvereport", tollbooth.LimitFuncHandler(lmt, controllers.ResolveReport)).Methods("POST")
	s.Handle("/serializereports", tollbooth.LimitFuncHandler(lmt, q.SerializeReports)).Methods("GET")

	//admin routes
	s.Handle("/createmod", tollbooth.LimitFuncHandler(lmt, controllers.CreateMod)).Methods("POST")
	return router
}
