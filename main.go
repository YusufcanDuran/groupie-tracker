package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

type GroupInformation struct {
	ID            int                 `json:"id"`
	Image         string              `json:"image"`
	Name          string              `json:"name"`
	Members       []string            `json:"members"`
	CreationDate  int                 `json:"creationDate"`
	FirstAlbum    string              `json:"firstAlbum"`
	DateLocations map[string][]string `json:"datesLocations"`
}

type Relation struct {
	DateLocations map[string][]string `json:"datesLocations"`
}

func main() {
	http.Handle("/templates/styles.css", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.HandleFunc("/", handleGroups)
	http.HandleFunc("/temp", handleind)

	// start web server
	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Something went wrong!")
		return
	}
}

var groups []GroupInformation

func handleGroups(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		fmt.Println("HTTP Get hatasi:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("HTTP yaniti okuma hatasi:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// var groups []GroupInformation
	err = json.Unmarshal(body, &groups)

	if err != nil {
		fmt.Println("JSON çözme hatasi:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Her bir grup için ilişkileri çek
	for i, group := range groups {
		relations, err := getRelationsByID(group.ID) // group.ID = her bir group verisi içindeki id alanlarını temsil eder
		if err != nil {                              // bu kısımda işlemler group içinde id 'nin temsil ettiği yerleri çekmeye çalışıyor getRelationsByID fonksiyonunun çağırılmasının temel nedeni, her bir grup yapısının içinde bulunan Relation yapısının ID alanını çekmek içindir.
			fmt.Printf("İlişkileri çekerken hata oluştu: %v\n", err)
			continue
		}
		// Grubun ilişkilerini GroupInformation yapısına ekle GroupInformation yapısının DateLocations alanına ekleniyor
		groups[i].DateLocations = relations.DateLocations // Relation yapısal türünde bulunan datesLocations alanındaki verileri, GroupInformation yapısal türünde bulunan DateLocations alanına atar.
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("HTML şablonu oluşturma hatasi:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, groups)
	if err != nil {
		fmt.Println("HTML çıktısı gönderme hatası:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	// Detaylı bilgi sayfasına yönlendirme yapacak olan HTTP işlemleri
}

func handleind(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	// İlgili grup ID'sini bul

	var group GroupInformation
	for _, g := range groups {
		if strconv.Itoa(g.ID) == id {
			group = g
			break
		}
	}
	if group.ID == 0 {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	// Grubun detaylı bilgilerini şablon dosyasına gönder
	tmpl, err := template.ParseFiles("templates/main.html")
	if err != nil {
		fmt.Println("HTML şablonu oluşturma hatasi:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, group)
	if err != nil {
		fmt.Println("HTML çıktısı gönderme hatası:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func getRelationsByID(id int) (*Relation, error) {
	resp, err := http.Get(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var relations Relation
	err = json.Unmarshal(body, &relations)
	if err != nil {
		return nil, err
	}

	return &relations, nil
}
