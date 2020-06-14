package main

import (
	"crypto/md5"
	"encoding/json"
	"log"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/0zAND1z/ipldcrud"
	"github.com/TexaProject/texajson"
	"github.com/TexaProject/texalib"
)

// AIName exports form value from /welcome globally
var AIName string

// IntName exports form value from /texa globally
var IntName string

// // Results is a collection of result from each AI
// type Results struct {
// 	Results []Result
// }

// var allResultsData Results

// Result is the master table recording the results of all test sessions for a given AI
type Result struct {
	AIName         string          `json:"AIName"`
	Interrogations []Interrogation `json:"Interrogations"`
}

// Interrogation is used to record the data from a session
type Interrogation struct {
	IntName  string                 `json:"IntName"`
	ArtiMts  float64                `json:"ArtiMts"`
	HumanMts float64                `json:"HumanMts"`
	CatVal   []texajson.CatValArray `json:"CatVal"`
}

// // CatVal is a structure to record the SPF for each slab or category
// type CatVal struct {
// 	CatName string `json:"CatName"`
// 	Spf     int    `json:"Spf"`
// }

// // IsResultObjectExists is used to check if Result object exists for the given AI's name
// func (allResultsData Results) IsResultObjectExists(aiName string) bool {
// 	// _, ok := allResultsData[aiName]
// 	// if ok {
// 	// 	return true
// 	// }
// 	// return false
// 	for _, result := range allResultsData.Results {
// 		if result.AIName == aiName {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (allResultsData Results) LoadResultObject(aiName string) (Result, bool) {
// 	for _, result := range allResultsData.Results {
// 		if result.AIName == aiName {
// 			return result, true
// 		}
// 	}
// 	return Result{}, false
// }

// // AddInterrogationData is used to update the new session data of an existing AI
// func (ResultsData Result) AddInterrogationData(newSessionData Interrogation) {
// 	ResultsData.Interrogations = append(ResultsData.Interrogations, newSessionData)
// 	return
// }

// NewResultObject is used to create a new Result object for a new AI
func NewResultObject(aiName string) Result {
	return Result{
		AIName:         aiName,
		Interrogations: []Interrogation{},
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/welcome", 301)
}

func texaHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// fmt.Printf("%+v\n", r.Form)
		fmt.Fprint(w, "<html><head><link rel=\"stylesheet\" href=\"http://localhost:3030/css/bootstrap.min.css\"><title>File Ack | TEXA Project</title></head><body>ACKNOWLEDGEMENT: Received the scores. <br /><br />Info:<br />")
		fmt.Fprint(w, "<br /><br />VISIT: /result for interrogation.")
		fmt.Fprintf(w, "<br /><br /><input type=\"button\" class=\"btn info\" onclick=\"location.href='http://localhost:3030/result';\" value=\"Visit /result\" /></body></html>")

		fmt.Println("--INTERROGATION FORM DATA--")
		IntName = r.Form.Get("IntName")
		QSA := r.Form.Get("scoreArray")
		SlabName := r.Form.Get("SlabName")
		slabSequence := r.Form.Get("slabSequence")

		fmt.Println("###", AIName)
		fmt.Println("###", IntName)
		fmt.Println("###", QSA)
		fmt.Println("###", SlabName)
		fmt.Println("###", slabSequence)

		// LOGIC
		re := regexp.MustCompile("[0-1]+")
		array := re.FindAllString(QSA, -1)

		SlabNameArray := regexp.MustCompile("[,]").Split(SlabName, -1)
		slabSeqArray := regexp.MustCompile("[,]").Split(slabSequence, -1)

		fmt.Println("###Resulting Array:")
		for x := range array {
			fmt.Println(array[x])
		}

		fmt.Println("###SlabNameArray: ")
		fmt.Println(SlabNameArray)

		fmt.Println("###slabSeqArray: ")
		fmt.Println(slabSeqArray)

		ArtiQSA := texalib.Convert(array)
		fmt.Println("###ArtiQSA:")
		fmt.Println(ArtiQSA)

		HumanQSA := texalib.SetHumanQSA(ArtiQSA)
		fmt.Println("###HumanQSA:")
		fmt.Println(HumanQSA)

		TSA := texalib.GetTransactionSeries(ArtiQSA, HumanQSA)
		fmt.Println("###TSA:")
		fmt.Println(TSA)

		// The Mean Test Scores(MTS) help us understand who performed better in the test
		ArtiMts := texalib.GetMeanTestScore(ArtiQSA)
		HumanMts := texalib.GetMeanTestScore(HumanQSA)

		fmt.Println("###ArtiMts: ", ArtiMts)
		fmt.Println("###HumanMts: ", HumanMts)

		PageArray := texajson.GetPages()
		fmt.Println("###PageArray")
		fmt.Println(PageArray)
		for _, p := range PageArray {
			fmt.Println(p)
		}

		newPage := texajson.ConvtoPage(AIName, IntName, ArtiMts, HumanMts)

		PageArray = texajson.AddtoPageArray(newPage, PageArray)
		fmt.Println("###AddedPageArray")
		fmt.Println(PageArray)

		JsonPageArray := texajson.ToJson(PageArray)
		fmt.Println("###jsonPageArray:")
		fmt.Println(JsonPageArray)

		////
		fmt.Println("### SLAB LOGIC")

		slabPageArray := texajson.GetSlabPages()
		fmt.Println("###slabPageArray")
		fmt.Println(slabPageArray)

		slabPages := texajson.ConvtoSlabPage(ArtiQSA, SlabNameArray, slabSeqArray)
		fmt.Println("###slabPages")
		fmt.Println(slabPages)
		for z := 0; z < len(slabPages); z++ {
			slabPageArray = texajson.AddtoSlabPageArray(slabPages[z], slabPageArray)
		}
		fmt.Println("###finalslabPageArray")
		fmt.Println(slabPageArray)

		JsonSlabPageArray := texajson.SlabToJson(slabPageArray)
		fmt.Println("###JsonSlabPageArray: ")
		fmt.Println(JsonSlabPageArray)

		////
		fmt.Println("### CAT LOGIC")

		CatPageArray := texajson.GetCatPages()
		fmt.Println("###CatPageArray")
		fmt.Println(CatPageArray)

		CatPages := texajson.ConvtoCatPage(AIName, slabPageArray, SlabNameArray)
		fmt.Println("###CatPages")
		fmt.Println(CatPages)
		CatPageArray = texajson.AddtoCatPageArray(CatPages, CatPageArray)

		// for z := 0; z < len(CatPages); z++ {
		// 	CatPageArray = texajson.AddtoCatPageArray(CatPages[z], CatPageArray)
		// }
		fmt.Println("###finalCatPageArray")
		fmt.Println(CatPageArray)

		JsonCatPageArray := texajson.CatToJson(CatPageArray)
		fmt.Println("###JsonCatPageArray: ")
		fmt.Println(JsonCatPageArray)

		ResultObject := NewResultObject(AIName)

		newSessionData := NewInterrogationObject(IntName, ArtiMts, HumanMts, CatPages.CatVal)
		fmt.Println("PRINTING NEW SESSION DATA BEFORE ADDING: ", newSessionData)

		ResultObject.Interrogations = append(ResultObject.Interrogations, newSessionData)
		fmt.Println("PRINTING UPDATED RESULT OBJECT: ", ResultObject)

		// fmt.Println("FINAL DATA IN BYTES: ", finalData)
		// WriteToLocalCache(finalData)
		cid := WriteDataToIPFS(ResultObject)
		if len(cid) > 0 {
			fmt.Println("Successfully wrote the session data to IPFS at ", cid)
		}
	}
}

// NewInterrogationObject is created a new object and returns it
func NewInterrogationObject(IntName string, ArtiMts, HumanMts float64, CatVal []texajson.CatValArray) Interrogation {
	return Interrogation{
		IntName:  IntName,
		ArtiMts:  ArtiMts,
		HumanMts: HumanMts,
		CatVal:   CatVal,
	}
}

// WriteDataToIPFS is used to write a data to IPFS using ipldcrud and return the CID
func WriteDataToIPFS(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("WriteDataToIPFS(): Issue in marshaling data!")
	}
	sh := ipldcrud.InitShell("http://localhost:5001") // Can be replaced with any hosted IPFS API URL also. Example: https://ipfs.infura.io:5001
	resultCid := ipldcrud.Set(sh, bytes)
	fmt.Println("WriteDataToIPFS(): Results of this testing session are globally accessible at https://explore.ipld.io/#/explore/" + resultCid)
	fmt.Println("WriteDataToIPFS(): You can also access them locally through ipld-explorer at http://localhost:3000/#/explore/" + resultCid)
	return resultCid
}

// func LoadFromLocalCache() Results {
// 	RedisClient := redis.NewClient(&redis.Options{
// 		Addr: "127.0.0.1:6379",
// 	})
// 	result, err := RedisClient.Ping().Result()
// 	if err != nil {
// 		panic("Err Connecting to Redis")
// 	} else {
// 		fmt.Println("Connected to Redis", result)
// 	}

// 	raw, err := RedisClient.Get("results").Result()
// 	if err != nil && err.Error() != "redis: nil" {
// 		fmt.Println(err.Error())
// 		os.Exit(1)
// 	}

// 	var finalDataFromCache Results
// 	json.Unmarshal([]byte(raw), &finalDataFromCache)
// 	return finalDataFromCache
// }

// func WriteToLocalCache(finalData []byte) {
// 	RedisClient := redis.NewClient(&redis.Options{
// 		Addr: "127.0.0.1:6379",
// 	})
// 	result, err := RedisClient.Ping().Result()
// 	if err != nil {
// 		panic("Err Connecting to Redis")
// 	} else {
// 		fmt.Println("Connected to Redis", result)
// 	}

// 	err = RedisClient.Set("results", string(finalData), 0).Err()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		os.Exit(1)
// 	}
// }

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/welcome.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
	}
}

// upload logic
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("login.html")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		handler.Filename = "elizadata.js"
		AIName = r.FormValue("AIName")
		fmt.Println(AIName)
		defer file.Close()

		fmt.Fprint(w, "<html><head><link rel=\"stylesheet\" href=\"http://localhost:3030/css/bootstrap.min.css\"><title>File Ack | TEXA Project</title></head><body>ACKNOWLEDGEMENT: Uploaded the file. <br /><br />Header Info:<br />")
		fmt.Fprintf(w, "%v", handler.Header)
		fmt.Fprintf(w, "<br /><br />Saved As: www/js/"+handler.Filename)
		fmt.Fprint(w, "<br /><br />VISIT: /texa for interrogation.")
		fmt.Fprintf(w, "<br /><br /><input type=\"button\" class=\"btn info\" onclick=\"location.href='http://localhost:3030/texa';\" value=\"Visit /texa\" /></body></html>")
		f, err := os.OpenFile("./www/js/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Selected file: ", handler.Filename)
		defer f.Close()
		io.Copy(f, file)
		// http.Redirect(w, r, "/texa", 301)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/result.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
	}
}

func getCatJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	catPages := texajson.GetCatPages()
	bs, err := json.Marshal(catPages)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func getMtsJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	mtsPage := texajson.GetPages()
	bs, err := json.Marshal(mtsPage)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func getSlabJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	slabPages := texajson.GetSlabPages()
	bs, err := json.Marshal(slabPages)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func main() {
	fmt.Println("--TEXA SERVER--")
	fmt.Println("STATUS: INITIATED")
	fmt.Println("ADDR: http://127.0.0.1:3030")

	fsc := http.FileServer(http.Dir("www/css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsc))
	fsj := http.FileServer(http.Dir("www/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsj))
	fsd := http.FileServer(http.Dir("www/data"))
	http.Handle("/data/", http.StripPrefix("/data/", fsd))

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/texa", texaHandler)
	http.HandleFunc("/result", resultHandler)
	http.HandleFunc("/cat", getCatJSON)
	http.HandleFunc("/mts", getMtsJSON)
	http.HandleFunc("/slab", getSlabJSON)

	http.ListenAndServe(":3030", nil)
}
