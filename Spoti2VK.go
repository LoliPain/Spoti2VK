package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)


const SpotiAuthCode = ""

// Как получить код первичной авторизации? Переходите по ссылке ниже:
// https://accounts.spotify.com/authorize?client_id=9f8b578a260549ab8338cf6263396975&response_type=code&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=user-read-currently-playing
// Входите в свой аккаунт Spotify внутри OAuth
// Принимаете условия авторизации приложения Spoti2VK OAuth
// В открывшейся ссылке, ищем code=
// Копируем все символы идущие после code и до конца строки. Это и будет ваш SpotiAuthCode

const VKToken  = ""

// Для получения валидного токена, для работы с API методами категории audio, нам понадобится приложение, имеющее к нему доступ
// Я использую токен от Kate Mobile, который получить можно здесь:
// https://vkhost.github.io
// Выбираете Kate Mobile в списке приложений, подтверждаете авторизацию через OAuth
// Скопируйте часть адресной строки от access_token= до &expires_in в открывшейся вкладке


type SpotifyInfo struct {
	SpotiToken string
	SpotyAuthCode string
	SpotiRefreshCode string
	SpotiNowPlaying string
}

func main(){
	var SpotiStruct SpotifyInfo
	SpotiStruct.SpotyAuthCode = SpotiAuthCode
	for true {
		SpotifyGetStatus(&SpotiStruct)
		time.Sleep(10*(time.Second))
	}
}

func SpotifyGetStatus(SpotiStruct *SpotifyInfo) {
	request, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", SpotiStruct.SpotiToken))
	client := &http.Client{}
	response, _ := client.Do(request)
	body, _ := ioutil.ReadAll(response.Body)
	var info interface{}
	_ = json.Unmarshal(body, &info)
	if info != nil {
		if info.(map[string]interface{})["error"] != nil {
			SpotiStruct.GetNewToken()
		} else if info.(map[string]interface{})["item"] != nil {
			ArtistRow := info.(map[string]interface{})["item"].(map[string]interface{})["album"].(map[string]interface{})["artists"].([]interface{})
			ArtistBuffer := bytes.Buffer{}
			for i := range ArtistRow {
				ArtistBuffer.WriteString(fmt.Sprintf("%s ", ArtistRow[i].(map[string]interface{})["name"].(string)))
			}
			Song := info.(map[string]interface{})["item"].(map[string]interface{})["name"].(string)
			FullTitle := strings.ReplaceAll(fmt.Sprintf("%s %s", ArtistBuffer.String(), Song), " ", "%20")
			if SpotiStruct.SpotiNowPlaying != FullTitle {
				GetVKMusic(FullTitle, strings.ReplaceAll(Song, " ", "%20"))
				SpotiStruct.SpotiNowPlaying = FullTitle
			}
		}
	}
}

func GetVKMusic(FullSpotifyTitle string, SongSpotifyOnly string) {
	VKFullResponse := VKQuery(FullSpotifyTitle)
	if VKFullResponse.(map[string]interface{})["count"].(float64) != 0 {
		Id := VKFullResponse.(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["id"].(float64)
		OwnerId := VKFullResponse.(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["owner_id"].(float64)
		SetVKStatus(fmt.Sprintf("%s_%s", strconv.Itoa(int(OwnerId)), strconv.Itoa(int(Id))))
	} else{
		VKTitleResponse := VKQuery(SongSpotifyOnly)
		if VKTitleResponse.(map[string]interface{})["count"].(float64) != 0 {
			Id := VKTitleResponse.(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["id"].(float64)
			OwnerId := VKTitleResponse.(map[string]interface{})["items"].([]interface{})[0].(map[string]interface{})["owner_id"].(float64)
			SetVKStatus(fmt.Sprintf("%s_%s", strconv.Itoa(int(OwnerId)), strconv.Itoa(int(Id))))
		}
	}
}

func SetVKStatus(AudioID string) {
	http.Get(fmt.Sprintf("https://api.vk.com/method/audio.setBroadcast?audio=%s&access_token=%s&v=5.1", AudioID, VKToken))
}

func (SpotiStruct *SpotifyInfo) GetNewToken() {
	r := url.Values{
	"grant_type" : {"authorization_code"},
	"redirect_uri" : {"https://example.com/callback"},
	"client_id" : {"9f8b578a260549ab8338cf6263396975"},
	"client_secret" : {"f27102cc5887410fb8250973cb0985ab"},
	"code" : {SpotiStruct.SpotyAuthCode},
	}
	resp, _ := http.PostForm("https://accounts.spotify.com/api/token",r)
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	var auth interface{}
	_ = json.Unmarshal(contents, &auth)
	var SpotiStructGet SpotifyInfo
	if auth.(map[string]interface{})["error"] != nil{
		if auth.(map[string]interface{})["error_description"] != "Authorization code expired"{
				panic(auth.(map[string]interface{})["error_description"])
		}else if SpotiStructGet.SpotiRefreshCode != ""{
			r := url.Values{
				"grant_type" : {"refresh_token"},
				"refresh_token" : {SpotiStruct.SpotiRefreshCode},
				"client_id" : {"9f8b578a260549ab8338cf6263396975"},
				"client_secret" : {"f27102cc5887410fb8250973cb0985ab"},
				"code" : {SpotiStruct.SpotyAuthCode},
			}
			resp, _ := http.PostForm("https://accounts.spotify.com/api/token",r)
			defer resp.Body.Close()
			contents, _ := ioutil.ReadAll(resp.Body)
			var auth interface{}
			_ = json.Unmarshal(contents, &auth)
			SpotiStruct.SpotiRefreshCode = auth.(map[string]interface{})["refresh_token"].(string)
			SpotiStruct.SpotiToken = auth.(map[string]interface{})["access_token"].(string)
		}else{
			panic("Authorization code expired and refresh code missing.\n Try to paste new SpotiAuthCode.")
		}
	} else{
		SpotiStruct.SpotiRefreshCode = auth.(map[string]interface{})["refresh_token"].(string)
		SpotiStruct.SpotiToken = auth.(map[string]interface{})["access_token"].(string)
	}
}

func VKQuery(SpotifyQuery string) interface{}{
	resp, _ := http.Get(fmt.Sprintf("https://api.vk.com/method/audio.search?q=%s&auto_complete=1&count=1&access_token=%s&v=5.1", SpotifyQuery, VKToken))
	body, _ := ioutil.ReadAll(resp.Body)
	var vkaudioget interface{}
	_ = json.Unmarshal(body, &vkaudioget)
	return vkaudioget.(map[string]interface{})["response"]
}