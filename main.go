package main

import (
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
	"encoding/json"

	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
)

type State struct {
	Latitude float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

func getCoords() (map[string]State, []string) {
	
	jsFile, err := os.Open("StateLatLon.json")
	if err != nil {
		fmt.Println("File Opening Error")
		fmt.Println(err)
		os.Exit(1)
	}
	//Closes the jsfile after the return
	defer jsFile.Close()

	//makes the js data readable for go
	jsData, err := ioutil.ReadAll(jsFile)
	if err != nil {
		fmt.Println("JSON Read Error")
		fmt.Println(err)
		os.Exit(1)
	}

	var states map[string]State

	//converts the js data into a map of the state struct with the keys being strings
	err = json.Unmarshal(jsData, &states)
	if err != nil {
		fmt.Println("JSON Unmarshal Error")
		fmt.Println(err)
		os.Exit(1)
	}

	//Creates a slice of the keys in the map to be used in the GUI
	keys := make([]string, len(states))
	i := 0
	for k := range states {
    	keys[i] = k
    	i++
	}
	slices.Sort(keys)

	return states, keys
}

func apiCall(latitude string, longitude string) string {
	
	//weather.gov doesn't have any api calls for specific states, closest is a lat & long option which then gives recommendations for other api calls with actual forecast info
	response, err := http.Get("https://api.weather.gov/points/"+latitude+","+longitude)
	
	//err / error is a predeclared types, and nil is a kind of zero value that goes with it and others when there's no data. Kinda like null, kinda not
	if err != nil {
		fmt.Println("API Call Error Occured:")
		fmt.Println(err.Error())
		os.Exit(1)
	} 
	
	//Reads the response data and puts it into a json format understandable by go
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Read Response Error Occured:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//Turns the json response data into a map of maps
	var responseMap map[string]interface{}
	err = json.Unmarshal([]byte(responseData), &responseMap)
	if err != nil {
		fmt.Println("Unmarshal Error Occured:")
		fmt.Println(err)
	}

	//Parses the response map of maps for the api call I really need, the forecast for the points that correlated to the lat and long inputted before (not the same values)
	// var nestedMap = responseMap["properties"].(map[string]interface{})["forecast"].(string)
	// var newCall = nestedMap["forecast"].(string)
	var newCall = responseMap["properties"].(map[string]interface{})["forecast"].(string)

	newResponse, err := http.Get(newCall)
	if err != nil {
		fmt.Println("New Call Error Occured:")
		fmt.Println(err)
	}

	newResponseData, err := ioutil.ReadAll(newResponse.Body)
	if err != nil {
		fmt.Println("Read New Response Error Occured:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var newResponseMap map[string]interface{}
	err = json.Unmarshal([]byte(newResponseData), &newResponseMap)
	if err != nil {
		fmt.Println("Unmarshal Error Occured:")
		fmt.Println(err)
	}

	//Essentially the same thing done above, just a few more times, including one indexing to an array of interface (go json really is something)
	return newResponseMap["properties"].(map[string]interface{})["periods"].([]interface{})[0].(map[string]interface{})["detailedForecast"].(string)

}

func apiPurser(test string) string {
	return test
}

func main() {
	
	stateCoords, keys := getCoords()

	mainApp := app.NewWithID("Golang Weather API GUI")
	mainWindow := mainApp.NewWindow("Golang Weather API GUI")
	mainWindow.Resize(fyne.NewSize(480, 240))

	//See Data Binding -> Two-Way Binding in documentation
	//Essentially allows the string called by the secondary label to be updated by the button so the API call doesn't go out until the state is selected
	bindingStr := binding.NewString()

	header := widget.NewLabel("Please Select State:")

	//see Widgets -> Choices in documentation
	combo := widget.NewSelect(keys, func(value string) {
	})

	//see Widgets -> ProgressBar in documentation
	loading := widget.NewProgressBarInfinite()
	loading.Hide()

	button := widget.NewButton("Get Forecast", func() {
		if combo.Selected != "" {
			loading.Show()
			//Do the API call in another thread to avoid Fyne freezing for that time.
			go func() {
				bindingStr.Set(apiCall(fmt.Sprint(stateCoords[combo.Selected].Latitude), fmt.Sprint(stateCoords[combo.Selected].Longitude)))
				//The loading bar only disappears after the API call has finished
				fyne.Do(func() {
					loading.Hide()
				})
			}()
		} else {
			bindingStr.Set("Please select state")
		}
	})

	secondary := widget.NewLabelWithData(bindingStr)
	secondary.Wrapping = fyne.TextWrapWord
	//Only used during testing to be able to scroll to see the entire JSON. Left as reference or in case the forecast is huge for some reason
	scrollContainer := container.NewScroll(secondary)
	scrollContainer.SetMinSize(fyne.NewSize(480, 240))

	mainWindow.SetContent(container.NewVBox(header, combo, button, loading, scrollContainer))
	mainWindow.ShowAndRun()

}