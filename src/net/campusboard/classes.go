package campusboard

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"schoolbot/internal/model"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var siteURL string = "https://campusboard.hs-kl.de/portalapps/sp/Semesterplan.do"

type major struct {
	Name string
}

type class struct {
	Name   string
	Abbr   string
	Sem    string
	Majors map[string]*major
}

func UpdateDB(box *model.ClassBox) (int, error) {
	clsList, err := scanAllClasses()
	if err != nil {
		return 0, err
	}
	n := 0
	query := box.Query(model.Class_.Name.Equals("", false))
	query.Limit(1)
	for _, cls := range clsList {

		query.SetStringParams(model.Class_.Name, cls.Name)
		results, err := query.Find()
		if err != nil {
			return 0, err
		} else if len(results) == 0 { //i.e. the class does not yet exist in the DB

			majList := make([]string, 0, 5)

			//Parse map of Major objects into list of just Major names (keys)
			for k := range cls.Majors {
				majList = append(majList, k)
			}

			box.Put(&model.Class{
				Name:   cls.Name,
				Abbr:   cls.Abbr,
				Majors: majList,
			})
			n++
		}
	}

	return n, nil
}

//Scans Campusboard for all classes of all majors and semesters
func scanAllClasses() (map[string]*class, error) {

	masterList := make(map[string]*class)

	client := &http.Client{}

	val := url.Values{}
	val.Set("action", "view")
	val.Set("studiengang", "460")
	val.Set("studsem", "1")
	val.Set("stundenplan", "ag")
	val.Set("anzeigeArt", "alle")

	t := time.Now()
	year := t.Year()
	month := t.Month()
	var sem string
	switch {
	case month < 3:
		sem = fmt.Sprintf("%d2", year-1)
	case month < 9:
		sem = fmt.Sprintf("%d1", year)
	default:
		sem = fmt.Sprintf("%d2", year)
	}
	val.Set("semester", sem)

	site, err := postFormToStr(client, val)
	if err != nil {
		return make(map[string]*class), err
	}

	majors := getMajors(site)
	for k, v := range majors {
		//create major object from map returned by GetMajors
		thisMajor := major{k} //It's ugly retrofitting but it works.
		fmt.Printf("Scanning major %v\n", k)
		val.Set("studiengang", v)
		site, err = postFormToStr(client, val)
		if err != nil {
			return make(map[string]*class), err
		}
		sems := getSemseters(site)
		for ks, vs := range sems {
			fmt.Println(" | " + ks)
			val.Set("studsem", vs)
			site, err = postFormToStr(client, val)
			if err != nil {
				return make(map[string]*class), err
			}
			classes := getClasses(site)
			for _, cl := range classes {
				//ignore classes that already exist in the  master list
				if masterList[cl.Abbr] == nil {
					masterList[cl.Abbr] = cl
					cl.Sem = vs
					cl.Majors = make(map[string]*major)
					cl.Majors[k] = &thisMajor
				} else if cl.Majors[k] == nil {
					masterList[cl.Abbr].Majors[k] = &thisMajor
				}
			}
		}
	}
	return masterList, nil
}

func postFormToStr(client *http.Client, val url.Values) (string, error) {
	resp, err := client.PostForm(siteURL, val)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return toUtf8(body), nil
}

//Extracts the available Majors from the given response
func getMajors(raw string) map[string]string {

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		log.Fatal(err)
	}

	//get select menu node with majors in it
	selNode := FindNode(doc, condAnd(condTypeElement(), condAnd(condHasData("select"), condHasAttr("name", "studiengang")))) // "select", "name", "studiengang")
	return SelectFieldToMap(selNode)

}

//Extracts the available semesters from the given response
func getSemseters(raw string) map[string]string {

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		log.Fatal(err)
	}

	//get select menu node with semesters in it
	selNode := FindNode(doc, condAnd(condTypeElement(), condAnd(condHasData("select"), condHasAttr("name", "studsem"))))
	return SelectFieldToMap(selNode)

}

//Extracts the available classes from the given response
func getClasses(raw string) []*class {

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		log.Fatal(err)
	}

	classList := make([]*class, 0, 15)

	//get table cells. Each table cell with the class="normal_splan" property represents a time slot that contains at least one class
	timeCells := FindAll(doc, condAnd(condTypeElement(), condAnd(condHasData("td"), condHasAttr("class", "normal_splan"))))

	for _, tCell := range timeCells {
		//in each time cell, each <table> element represents a class time slot
		tSlots := FindAll(tCell, condAnd(condTypeElement(), condHasData("table")))

		for _, tSlot := range tSlots {
			class := class{}
			//find all tr nodes in the class
			tRows := FindAll(tSlot, condAnd(condTypeElement(), condHasData("tr")))
			i := 0
			for _, n := range tRows {
				//find the first td node (it's the only one that matters)
				if i == 2 {
					td := FindNode(n, condAnd(condTypeElement(), condHasData("td")))
					abbr := FindNode(td, condTypeText()).Data
					abbr = strings.TrimSpace(abbr)
					abbr = strings.ReplaceAll(abbr, ";", "")
					class.Abbr = strings.Split(abbr, "-")[0] //removes extentions from abbreviation
					//find title attribute in tr node and get the class name
					for _, att := range td.Attr {
						if att.Key == "title" {
							name := strings.ReplaceAll(att.Val, ";", "")
							name = strings.TrimSpace(name)
							regex := regexp.MustCompile("(([vV]orlesung)|([üÜ]bung))|\\((V|Ü|v|ü)+\\)|-?(V|Ü|v|ü)")
							class.Name = regex.ReplaceAllString(name, "")
							break
						}
					}
				}
				i++
			}
			classList = append(classList, &class)
		}
	}
	return classList
}
