package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ihleven/ihle.cloud/backend/auth"
	"github.com/ihleven/ihle.cloud/backend/hi"
)

func FileHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, "missing cookie", 401)
		return
	}
	claims, err := auth.AuthenticatorPKG.ParseClaims(cookie.Value)
	if err != nil {
		http.Error(w, "missing claims", 401)
		return
	}

	account, accesstoken, err := GetAccessFromClaims(claims)

	// path := r.PathValue("path")
	// if i := strings.Index(r.PathValue("path"), "/"); i > 0 {

	// 	path = path[0:i]
	// }
	// for i, p := range claims.Permissions {

	// }

	fmt.Println("path:", r.PathValue("path"), account, claims)
	perm := getPermissionForPath(r.PathValue("path"))
	if _, ok := claims.Permissions[perm]; !ok {
		http.Error(w, "missing permission "+perm, 403)
		return
	}

	resp, err := hi.NewClient(accesstoken, "").File(r.PathValue("path"), r.URL.Query(), r.Header)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Content-Disposition", "inline")

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)

	if err != nil {
		http.Error(w, "io error", 500)
	}
}

func MetaHandler(w http.ResponseWriter, r *http.Request) error {

	cookie, err := r.Cookie("jwt")
	if err != nil {
		return err
	}
	claims, err := auth.AuthenticatorPKG.ParseClaims(cookie.Value)
	if err != nil {
		return err
	}
	fmt.Printf("CLAIMS: %v\n", claims)
	_, accesstoken, err := GetAccessFromClaims(claims)
	if err != nil {
		return err
	}
	fmt.Printf("accesstoken: %v\n", accesstoken)

	perm := getPermissionForPath(r.PathValue("path"))
	if _, ok := claims.Permissions[perm]; !ok {
		// http.Error(w, "missing permission "+perm, 403)
		return err
	}
	fmt.Printf("perm: %v\n", perm)

	client := hi.NewClient(accesstoken, "/")

	path := "/" + r.PathValue("path")

	var meta *hi.Meta

	if path == "" || strings.HasSuffix(path, "/") {
		meta, err = client.GetDir(path)
	} else {
		meta, err = client.GetMeta(path)
	}
	if err != nil {
		// http.Error(w, "meta error: "+err.Error(), 500)
		return err
	}

	return RespondJSON(w, meta)
}
