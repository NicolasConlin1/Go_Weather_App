# Go Weather App 
A basic weather app utilizing API calls to weather.gov made entirely in Golang with the Fyne GUI Library.

A modified version of an open source JSON document for the general latitude and longitude of each state is used to make an API call to weather.gov.
This information, provided in JSON, is then used to get the information needed for a more useful call back to weather.gov again for a forecast of the area.

A GUI made using the open source Fyne library allows the user to select their desired state before displaying the forecast.
As the API call can occasionally take a moment, multi-threading is used to display a loading bar and keep the GUI functioning during that time.
