 package rcp

import "time"




type WorkStatus struct {
    Status  string `json:"status"`
	Data string `json:"data"`

}

type WorkStart struct {
	Status  string `json:"status"`
	Data string `json:"data"`
}

type WorkStop struct {
    Status  string `json:"status"`
}


type WorkTimesheet struct {
    ID  string `json:"id"`
 UserID  string `json:"userid"`
	Startedat  string `json:"startedat"`
 Endedat  string `json:"endedat"`
 Minutes  string `json:"minutes"`


}

type DailyTotal struct {
	UserID          int
	Day             time.Time
	WorkedMinutes   int
	BreakMinutes    int
	OvertimeMinutes int
	AbsenceMinutes  int
}
