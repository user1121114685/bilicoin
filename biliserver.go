package bilicoin

import (
	"context"
	"github.com/gorilla/mux"
	. "github.com/r3inbowari/zlog"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/wuwenbao/gcors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var BiliServer *http.Server

func BCApplication() {

	Log.Info("[BCS] BILICOIN api Mode running")
	reset()
	Log.Info("[BCS] Listened on " + GetConfig(false).APIAddr)
	r := mux.NewRouter()

	r.HandleFunc("/{uid}/ft", HandleFT)
	r.HandleFunc("/{uid}/cron", HandleCron)
	r.HandleFunc("/version", HandleVersion)
	r.HandleFunc("/users", HandleUsers)
	r.HandleFunc("/user", HandleUserAdd).Methods("POST")
	r.HandleFunc("/user", HandleUserDel).Methods("DELETE")

	// allow CORS
	cors := gcors.New(
		r,
		gcors.WithOrigin("*"),
		gcors.WithMethods("POST, GET, PUT, DELETE, OPTIONS"),
		gcors.WithHeaders("Authorization"),
	)

	BiliServer = &http.Server{
		Addr:    GetConfig(false).APIAddr,
		Handler: cors,
	}
	err := BiliServer.ListenAndServe()
	//err := http.ListenAndServe(GetConfig(false).APIAddr, cors)
	if strings.HasSuffix(err.Error(), "normally permitted.") || strings.Index(err.Error(), "bind") != -1 {
		Log.WithFields(logrus.Fields{"err": err.Error()}).Fatal("[BCS] Only one usage of each socket address is normally permitted.")
		Log.Info("[BCS] EXIT 1002")
		os.Exit(1002)
	}

	// goroutine block here not need sleep
	Log.WithFields(logrus.Fields{"err": err.Error()}).Info("[BCS] Service will be terminated soon")
	time.Sleep(time.Second * 10)
}

func Shutdown(ctx context.Context) {
	if BiliServer != nil {
		Log.Info("[BSC] releasing server now...")
		err := BiliServer.Shutdown(ctx)
		if err != nil {
			Log.Fatal("[BSC] Shutdown failed")
			Log.Info("[BCS] EXIT 1002")
			os.Exit(1011)
		}
		Log.Info("[BSC] release completed")
	}
}

type FilterBiliUser struct {
	UID       string `json:"uid"`
	Cron      string `json:"cron"`
	FT        string `json:"ft"`
	FTSwitch  bool   `json:"ftSwitch"`
	DropCount int    `json:"dropCount"`
}

func HandleFT(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	vars := mux.Vars(r)
	uid := vars["uid"]
	key := r.Form.Get("key")
	sw := r.Form.Get("sw")

	if biu, ok := GetUser(uid); ok == nil && biu != nil {
		if key != "" {
			biu.FT = key
			biu.FTSwitch = true
		}
		if sw == "0" {
			biu.FTSwitch = false
		} else {
			biu.FTSwitch = true
		}
		biu.InfoUpdate()
		Log.WithFields(logrus.Fields{"UID": uid, "Key": key}).Info("[BCS] FTQQ secret key save completed")
	}
	ResponseCommon(w, "try succeed", "ok", 1, http.StatusOK, 0)
}

func HandleCron(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	vars := mux.Vars(r)
	uid := vars["uid"]
	cronStr := r.Form.Get("spec")

	if biu, ok := GetUser(uid); ok == nil && biu != nil {
		if _, err := cron.Parse(cronStr); err != nil {
			ResponseCommon(w, "[BCS] incorrect cron spec, please check and try again", "ok", 1, http.StatusOK, 0)
			Log.WithFields(logrus.Fields{"UID": uid, "Cron": cronStr}).Info("[BCS] incorrect cron spec, please check and try again")
			return
		}
		biu.Cron = cronStr
		biu.InfoUpdate()
		Log.WithFields(logrus.Fields{"UID": uid, "Cron": cronStr}).Info("[BCS] Cron save completed by web")
	}
	reset()
	ResponseCommon(w, "try succeed", "ok", 1, http.StatusOK, 0)
}

func HandleVersion(w http.ResponseWriter, r *http.Request) {
	ResponseCommon(w, releaseVersion+" "+releaseTag, "ok", 1, http.StatusOK, 0)
}

func HandleUsers(w http.ResponseWriter, r *http.Request) {
	users := LoadUser()
	var ret []FilterBiliUser
	for _, v := range users {
		ret = append(ret, FilterBiliUser{
			UID:       v.DedeUserID,
			Cron:      v.Cron,
			FT:        v.FT,
			FTSwitch:  v.FTSwitch,
			DropCount: v.DropCoinCount,
		})
	}
	ResponseCommon(w, ret, "ok", len(users), http.StatusOK, 0)
}

var loginMap sync.Map

func HandleUserAdd(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if r.Form.Get("oauth") == "" {
		// 提供回调
		user, _ := CreateUser()
		_ = user.GetQRCode()
		Log.WithFields(logrus.Fields{"oauth": user.OAuth.OAuthKey}).Info("[BCS] qrcode created")
		loginMap.Store(user.OAuth.OAuthKey, user)
		time.AfterFunc(time.Minute*3, func() {
			loginMap.Delete(user.OAuth.OAuthKey)
		})
		ResponseCommon(w, user.OAuth.OAuthKey, "ok", 1, http.StatusOK, 0)
	} else {
		if user, ok := loginMap.Load(r.Form.Get("oauth")); ok {
			biliUser := user.(*BiliUser)
			biliUser.LoginCallback(func(isLogin bool) {
				ResponseCommon(w, isLogin, "ok", 1, http.StatusOK, 0)
				if isLogin {
					// reset
					reset()
				}
			})
			// ResponseCommon(w, oauth, "ok", 1, http.StatusOK, 0)
		}
		//TODO not exist key
	}
}

func HandleUserDel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	uid := r.Form.Get("uid")
	Log.WithFields(logrus.Fields{"UID": uid}).Info("[BCS] Try to delete user")
	_ = DelUser(uid)
	reset()
	ResponseCommon(w, "try succeed", "ok", 1, http.StatusOK, 0)
}

func reset() {
	Log.Warn("[BSC] Release task resource")
	taskMap.Range(func(key, value interface{}) bool {
		Log.WithFields(logrus.Fields{"TaskID": key.(string)}).Info("[BSC] Try to stop cron")
		value.(*cron.Cron).Stop()
		taskMap.Delete(key)
		return true
	})
	CronTaskLoad()
}
