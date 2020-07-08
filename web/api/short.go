package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"net"

	"time"

	"shortme/conf"

	"crypto/md5"
	"encoding/hex"

	"shortme/short"

	"github.com/gorilla/mux"
)

func Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortededURL := vars["shortenedURL"]

	longURL, err := short.Shorter.Expand(shortededURL)
	if err != nil {
		log.Printf("redirect short url error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	} else {
		if len(longURL) != 0 {

			cookieName := shortededURL + "_Token"

			cookie, err := r.Cookie(cookieName)
			if err != nil {
				// 单位：秒。
				COOKIE_MAX_MAX_AGE := time.Hour * 24 / time.Second
				maxAge := int(COOKIE_MAX_MAX_AGE)
				cookieValue, ipAddr := GetCookieValue(r, longURL, shortededURL)
				uid_cookie := &http.Cookie{
					Name:     cookieName,
					Value:    cookieValue,
					Path:     "/",
					HttpOnly: false,
					MaxAge:   maxAge,
				}
				http.SetCookie(w, uid_cookie)
				//log.Println(uid_cookie)
				go short.Shorter.SaveAccessRecord(r.Header.Get("user-agent"), longURL, shortededURL, ipAddr)
			} else {
				log.Println(cookie.Value)
			}

			w.Header().Set("Location", longURL)
			w.WriteHeader(http.StatusTemporaryRedirect)

		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func md5V1(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GetCookieValue(r *http.Request, longURL, shortededURL string) (string, string) {
	publicIp := ""

	RemoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)

	if ipProxy := r.Header.Get("X-Forwarded-For"); len(ipProxy) > 0 {
		RemoteAddr = ipProxy
	} else if ipProxy = r.Header.Get("X-Real-Ip"); len(ipProxy) > 0 {
		RemoteAddr = ipProxy
	}
	if len(RemoteAddr) > 0 {
		publicIp = RemoteAddr
	}
	ua := r.Header.Get("user-agent")
	log.Println("user-agent:", ua)

	uuid := md5V1(ua + publicIp + longURL + shortededURL)
	log.Println("user-uuid:", uuid)
	return uuid, publicIp

}

func ShortURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read short request error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	}

	var shortReq shortReq
	err = json.Unmarshal(body, &shortReq)
	if err != nil {
		log.Printf("parse short request error. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	}

	var longURL *url.URL
	longURL, err = url.Parse(shortReq.LongURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: "requested url is malformed"})
		w.Write(errMsg)
		return
	} else {
		if longURL.Host == conf.Conf.Common.DomainName {
			w.WriteHeader(http.StatusBadRequest)
			errMsg, _ := json.Marshal(errorResp{Msg: "requested url is already shortened"})
			w.Write(errMsg)
			return
		}
		if strings.ToLower(longURL.Scheme) != "http" && strings.ToLower(longURL.Scheme) != "https" {
			w.WriteHeader(http.StatusBadRequest)
			errMsg, _ := json.Marshal(errorResp{Msg: "requested url is not a http or https url"})
			w.Write(errMsg)
			return
		}
	}

	var shortenedURL string
	shortenedURL, err = short.Shorter.Short(shortReq.LongURL)
	shortenedURL = (&url.URL{
		Scheme: conf.Conf.Common.Schema,
		Host:   conf.Conf.Common.DomainName,
		Path:   shortenedURL,
	}).String()
	if err != nil {
		log.Printf("short url error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	} else {
		shortResp, _ := json.Marshal(shortResp{ShortURL: shortenedURL})
		w.Write(shortResp)
	}
}

func ExpandURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read expand request error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	}

	var expandReq expandReq
	err = json.Unmarshal(body, &expandReq)
	if err != nil {
		log.Printf("parse expand request error. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	}

	var shortURL *url.URL
	shortURL, err = url.Parse(expandReq.ShortURL)
	if err != nil {
		log.Printf(`"%v" is not a valid url`, expandReq.ShortURL)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	} else {
		var expandedURL string
		expandedURL, err = short.Shorter.Expand(strings.TrimLeft(shortURL.Path, "/"))
		if err != nil {
			log.Printf("expand url error. %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
			w.Write(errMsg)
			return
		} else {
			expandResp, _ := json.Marshal(expandResp{LongURL: expandedURL})
			w.Write(expandResp)
		}
	}
}
