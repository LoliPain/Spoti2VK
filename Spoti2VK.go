package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/p2love/GoFrequency/GoFrequency"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const SpotiAuthCode = ""

// Как получить код первичной авторизации? Переходите по ссылке ниже:
// https://accounts.spotify.com/authorize?client_id=9f8b578a260549ab8338cf6263396975&response_type=code&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=user-read-currently-playing
// Входите в свой аккаунт Spotify внутри OAuth
// Принимаете условия авторизации приложения Spoti2VK OAuth
// В открывшейся ссылке, ищем code=
// Копируем все символы идущие после code и до конца строки. Это и будет ваш SpotiAuthCode

const VKToken = ""

// Для получения валидного токена, для работы с API методами категории audio, нам понадобится приложение, имеющее к нему доступ
// Я использую токен от Kate Mobile или  VK Admin, который получить можно здесь:
// https://vkhost.github.io
// Выбираете VK Admin (протестировано) в списке приложений, подтверждаете авторизацию через OAuth
// Скопируйте часть адресной строки от access_token= до &expires_in в открывшейся вкладке

const DebugMode = ""

// Включает/Выключает режим отладки, необходимый для поиска багов с поиском и включением треков
// Во включенном состоянии, перед тем, как попытаться поставить трек в статус, отправляет в консоль его название
// В случае паника, при попытке поставить статус, последняя запись в nohup.out будет (в большинстве случаев)
// Причиной его возникновения.
// Для включения режима отладки, необходимо указать const DebugMode = "True"

var vkUserId = GetVKId()

const SendDailyStatusReport = ""

// Включает/Выключает отправку NowPlaying для создания историй с последними прослушанными аудио в сервисе VK
// Для включения, необходимо указать const SendDailyStatusReport = "True"
// Ссылка на сервис:
// https://vk.com/app7334100

type SpotifyInfo struct {
	SpotiToken       string
	SpotiAuthCode    string
	SpotiRefreshCode string
	SpotiNowPlaying  string
}

func main() {
	var SpotiStruct SpotifyInfo
	SpotiStruct.SpotiAuthCode = SpotiAuthCode
	for true {
		SpotifyGetStatus(&SpotiStruct)
		time.Sleep(10 * (time.Second))
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
		info := info.(map[string]interface{})
		if info["error"] != nil {
			SpotiStruct.GetNewToken()
		} else if info["item"] != nil {
			ArtistRow := info["item"].(map[string]interface{})["album"].(map[string]interface{})["artists"].([]interface{})
			ArtistBuffer := bytes.Buffer{}
			for i := range ArtistRow {
				ArtistBuffer.WriteString(fmt.Sprintf("%s ", ArtistRow[i].(map[string]interface{})["name"].(string)))
			}
			Song := info["item"].(map[string]interface{})["name"].(string)
			FullTitle := fmt.Sprintf("%s %s", ArtistBuffer.String(), Song)
			if SpotiStruct.SpotiNowPlaying != FullTitle {
				if DebugMode == "True" {
					fmt.Println(FullTitle)
				}
				GetVKMusic(FullTitle, Song)
				if SendDailyStatusReport == "True" {
					SendDailyStatus(vkUserId)
				}
				SpotiStruct.SpotiNowPlaying = FullTitle
			}
		}
	}
}

func GetVKMusic(FullSpotifyTitle string, SongSpotifyOnly string) {
	VKFullResponse := VKQuery(FullSpotifyTitle)
	if VKFullResponse != nil {
		response := VKFullResponse.(map[string]interface{})
		if response["count"].(float64) != 0 && len(response["items"].([]interface{})) != 0 {
			SetVKStatus(SongChoose(response, FullSpotifyTitle))
		} else {
			VKTitleResponse := VKQuery(SongSpotifyOnly)
			response := VKTitleResponse.(map[string]interface{})
			if response["count"].(float64) != 0 && len(response["items"].([]interface{})) != 0 {
				SetVKStatus(SongChoose(response, FullSpotifyTitle))
			}
		}
	}
}

func SetVKStatus(AudioID string) {
	r := url.Values{
		"audio":        {AudioID},
		"access_token": {VKToken},
		"v":            {"5.1"},
	}
	_, _ = http.Get("https://api.vk.com/method/audio.setBroadcast?" + r.Encode())
}

func (SpotiStruct *SpotifyInfo) GetNewToken() {
	r := url.Values{
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {"https://example.com/callback"},
		"client_id":     {"9f8b578a260549ab8338cf6263396975"},
		"client_secret": {"f27102cc5887410fb8250973cb0985ab"},
		"code":          {SpotiStruct.SpotiAuthCode},
	}
	resp, _ := http.PostForm("https://accounts.spotify.com/api/token", r)
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	var auth interface{}
	_ = json.Unmarshal(contents, &auth)
	if auth.(map[string]interface{})["error"] != nil {
		if SpotiStruct.SpotiRefreshCode != "" {
			r := url.Values{
				"grant_type":    {"refresh_token"},
				"refresh_token": {SpotiStruct.SpotiRefreshCode},
				"client_id":     {"9f8b578a260549ab8338cf6263396975"},
				"client_secret": {"f27102cc5887410fb8250973cb0985ab"},
				"code":          {SpotiStruct.SpotiAuthCode},
			}
			resp, _ := http.PostForm("https://accounts.spotify.com/api/token", r)
			defer resp.Body.Close()
			contents, _ := ioutil.ReadAll(resp.Body)
			var auth interface{}
			_ = json.Unmarshal(contents, &auth)
			SpotiStruct.SpotiToken = auth.(map[string]interface{})["access_token"].(string)
		} else {
			panic(auth.(map[string]interface{})["error_description"])
		}
	} else {
		SpotiStruct.SpotiRefreshCode = auth.(map[string]interface{})["refresh_token"].(string)
		SpotiStruct.SpotiToken = auth.(map[string]interface{})["access_token"].(string)
	}
}

func VKQuery(SpotifyQuery string) interface{} {
	r := url.Values{
		"q":             {SpotifyQuery},
		"auto_complete": {"1"},
		"count":         {"10"},
		"access_token":  {VKToken},
		"v":             {"5.1"},
	}
	resp, _ := http.Get("https://api.vk.com/method/audio.search?" + r.Encode())
	body, _ := ioutil.ReadAll(resp.Body)
	var vkaudioget interface{}
	_ = json.Unmarshal(body, &vkaudioget)
	return vkaudioget.(map[string]interface{})["response"]
}

func SongChoose(response map[string]interface{}, status string) string {
	var score float64
	var set string
	for _, items := range response["items"].([]interface{}) {
		items := items.(map[string]interface{})
		responseTitle := fmt.Sprintf("%s %s", items["artist"], items["title"])
		if GoFrequency.Score(GoFrequency.Map(responseTitle), GoFrequency.Map(status)) > score {
			score = GoFrequency.Score(GoFrequency.Map(responseTitle), GoFrequency.Map(status))
			set = fmt.Sprintf("%f_%f", items["owner_id"].(float64), items["id"].(float64))
		}
	}
	return set
}

func GetVKId() string {
	k := url.Values{
		"access_token": {VKToken},
		"v":            {"5.100"},
	}
	vkGetUser, _ := http.Get("https://api.vk.com/method/users.get?" + k.Encode())
	body, _ := ioutil.ReadAll(vkGetUser.Body)
	var vkJsonUser interface{}
	var vkUserId float64
	_ = json.Unmarshal(body, &vkJsonUser)
	for _, g := range vkJsonUser.(map[string]interface{})["response"].([]interface{}) {
		vkUserId = g.(map[string]interface{})["id"].(float64)
	}
	return strconv.Itoa(int(vkUserId))
}

func SendDailyStatus(id string) {
	o := url.Values{
		"user_id": {id},
		"stamp":   {strconv.Itoa(int(time.Now().Unix()))},
	}
	_, _ = http.Get("https://worstin.me:5000/spoti2vk/api/add?" + o.Encode())
}
