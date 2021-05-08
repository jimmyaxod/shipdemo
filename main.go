package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

var speed float64

func ShipServer(ws *websocket.Conn) {

	replay_start_date, err := time.Parse(time.RFC3339, "2020-06-01T00:05:19+00:00")

	start_time := time.Now()

	if err != nil {
		panic(err)
	}

	// Read the data from file...
	file, err := os.Open("ship_data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		csvline := scanner.Text()
		// Split it
		bits := strings.Split(csvline, ",")
		// MMSI,BaseDateTime,LAT,LON,SOG,COG,Heading,VesselName,IMO,CallSign,VesselType,Status,Length,Width,Draft,Cargo,TranscieverClass
		lat, _ := strconv.ParseFloat(bits[2], 64)
		lon, _ := strconv.ParseFloat(bits[3], 64)
		id := bits[9]
		timestamp := bits[1]

		date, err := time.Parse(time.RFC3339, fmt.Sprintf("%s+00:00", timestamp))
		if err != nil {
			panic(err)
		}

		diff := date.Sub(replay_start_date)

		// TODO: Would be better to work out how long to sleep for...
		for {
			diff_replay := time.Since(start_time) * time.Duration(speed)
			if diff_replay > diff {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		status := fmt.Sprintf("Replaying %s", timestamp)

		data := fmt.Sprintf("%s,%f,%f,%s", id, lat, lon, status)
		fmt.Printf("SEND %s %s\n", timestamp, data)
		ws.Write([]byte(data))
	}
}

func main() {

	flag.Float64Var(&speed, "speed", 1000, "Speed to play the events")

	flag.Parse()

	log.Println("Starting webserver")
	// Serve static files...
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	http.Handle("/ws", websocket.Handler(ShipServer))

	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
