package main

import (
    "bufio"
    "fmt"
    "strconv"
    "os"
    "strings"
    "net/http"
    "sort"
    
    "github.com/labstack/echo/v4"
)

// 2023-08-05 13:25:58,547	name@provider.tld	939	279	19	user place
type Event struct {
    Time string
    User string
    X int
    Y int
    Color int
    Type string
}


// Type is either "user place" or "user undo"
const (
    UserPlace string = "user place"
    UserUndo string = "user undo"
)

type ReturnEvent struct {
    Time string
    User string
    X int
    Y int
    Color string
    Type string
}

type Color int


const (
    Black Color = iota
    DarkGrey
    DeepGrey
    MediumGrey
    LightGrey
    White
    Beige
    Peach
    Brown
    Chocolate
    Rust
    Orange
    Yellow
    PastelYellow
    Lime
    Green
    DarkGreen
    Forest
    DarkTeal
    LightTeal
    Aqua
    Azure
    Blue
    Navy
    Purple
    Mauve
    Magenta
    Pink
    Watermelon
    Red
    Rose
    Maroon
)

func (c Color) String() string {
    return [...]string{"Black", "Dark Grey", "Deep Grey", "Medium Grey", "Light Grey", "White", "Beige", "Peach", "Brown", "Chocolate", "Rust", "Orange", "Yellow", "Pastel Yellow", "Lime", "Green", "Dark Green", "Forest", "Dark Teal", "Light Teal", "Aqua", "Azure", "Blue", "Navy", "Purple", "Mauve", "Magenta", "Pink", "Watermelon", "Red", "Rose", "Maroon"}[c]
}

func parseLog() []Event {
    f, err := os.Open("pixels.log")
    if err != nil {
        fmt.Print("There has been an error!: ", err)
    }
    defer f.Close()

    lines := 0

    var events []Event

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines++
        
        line := scanner.Bytes()
        
        request := string(line)

        // split line into parts
        matches := strings.Split(request, "\t")

        // convert to int
        x, _ := strconv.Atoi(matches[2])
        y, _ := strconv.Atoi(matches[3])
        color, _ := strconv.Atoi(matches[4])

        event := Event{
            Time: matches[0],
            User: matches[1],
            X: x,
            Y: y,
            Color: color,
            Type: matches[5],
        }
        
        events = append(events, event)
    }

    if err := scanner.Err(); err != nil {
        fmt.Println(err)
    }

    return events
}


func getPixelHandler(c echo.Context) error {
    x, _ := strconv.Atoi(c.Param("x"))
    y, _ := strconv.Atoi(c.Param("y"))
    
    var results []ReturnEvent

    // search for pixel in log
    for _, event := range events {
        if event.X == x && event.Y == y {
            e := ReturnEvent{
                Time: event.Time,
                User: event.User,
                X: event.X,
                Y: event.Y,
                Color: Color(event.Color).String(),
                Type: event.Type,
            }
            results = append(results, e)
        }
    }

    if len(results) == 0 {
        return c.JSONPretty(http.StatusNotFound, "Pixel not found", "  ")
    } else {
        return c.JSONPretty(http.StatusOK, results, "  ")
    }
}

type PixelStats struct {
    X int
    Y int
    TimeChanged int
    FavouriteColor string
    MostActiveUser string
    Virgin bool
}

func getPixelStatsHandler(c echo.Context) error {
    x, _ := strconv.Atoi(c.Param("x"))
    y, _ := strconv.Atoi(c.Param("y"))
    
    var results []Event

    var colors []int

    var users []string

    // search for pixel in log
    for _, event := range events {
        if event.X == x && event.Y == y {
            colors = append(colors, event.Color)
            users = append(users, event.User)
            results = append(results, event)
        }
    }

    if len(results) != 0 {
        // calculate favourite color by doing the mean of all colors
        var sum int
        for _, color := range colors {
            sum += color
        }

        favouriteColor := Color(sum / len(colors)).String()

        // calculate most active user by doing the mode of all users
        counts := make(map[string]int)
        for _, user := range users {
            counts[user]++
        }

        var mostActiveUser string
        var mostActiveUserCount int
        for user, count := range counts {
            if count > mostActiveUserCount {
                mostActiveUser = user
                mostActiveUserCount = count
            }
        }


        stats := PixelStats{
            X: x,
            Y: y,
            TimeChanged: len(results),
            FavouriteColor: favouriteColor,
            MostActiveUser: mostActiveUser,
            Virgin: false,
        }

        return c.JSONPretty(http.StatusOK, stats, "  ")
    } else {
        stats := PixelStats{
            X: x,
            Y: y,
            TimeChanged: 0,
            FavouriteColor: "",
            MostActiveUser: "",
            Virgin: true,
        }

        return c.JSONPretty(http.StatusOK, stats, "  ")
    }
}

func getUserHandler(c echo.Context) error {
    user := c.Param("user")
    
    var results []ReturnEvent

    // search for user in log
    for _, event := range events {
        if event.User == user {
            e := ReturnEvent{
                Time: event.Time,
                User: event.User,
                X: event.X,
                Y: event.Y,
                Color: Color(event.Color).String(),
                Type: event.Type,
            }
            results = append(results, e)
        }
    }

    return c.JSONPretty(http.StatusOK, results, "  ")
}

type UserStats struct {
    User string
    PixelsPlaced int
    PixelsUndone int
    TotalPixels int
    FavouriteColor string
    PlaceInLeaderboard int
}

func getUserStatsHandler(c echo.Context) error {
    user := c.Param("user")
    
    var results []Event

    var colors []int

    pixelsplaced, pixelsundone, total := 0, 0, 0


    // search for user in log
    for _, event := range events {
        if event.User == user {
            colors = append(colors, event.Color)
            if event.Type == UserUndo {
                pixelsundone++
                total--
            } else {
                pixelsplaced++
                total++
                results = append(results, event)
            }
        }
    }

    // calculate favourite color by doing the mean of all colors
    var sum int
    for _, color := range colors {
        sum += color
    }


    favouriteColor := Color(sum / len(colors)).String()

    // calculate place in leaderboard
    leaderboard := makeLeaderBoard()

    placeInLeaderboard := 0

    for i, result := range leaderboard {
        if result.User == user {
            placeInLeaderboard = i + 1
        }
    }

    stats := UserStats{
        User: user,
        PixelsPlaced: pixelsplaced,
        PixelsUndone: pixelsundone,
        TotalPixels: total,
        FavouriteColor: favouriteColor,
        PlaceInLeaderboard: placeInLeaderboard,
    }

    return c.JSONPretty(http.StatusOK, stats, "  ")
}

type LeaderBoardUser struct {
    User string
    PixelsPlaced int
}

func makeLeaderBoard() []LeaderBoardUser {
    var results []LeaderBoardUser

    var users []string

    // search for user in log
    for _, event := range events {
        users = append(users, event.User)
    }

    // sort users by number of pixels placed
    counts := make(map[string]int)
    for _, user := range users {
        counts[user]++
    }

    for user, _ := range counts {
        stats := LeaderBoardUser{
            User: user,
            PixelsPlaced: counts[user],
        }
        results = append(results, stats)
    }


    sort.Slice(results, func(i, j int) bool {
        return results[i].PixelsPlaced > results[j].PixelsPlaced
    })
    return results
}

func getLeaderboard(c echo.Context) error {
    results := makeLeaderBoard()
    return c.JSONPretty(http.StatusOK, results[:10], "  ")
}

var events = parseLog()

func main() {
    e := echo.New()

    e.GET("/pixel/:x/:y", getPixelHandler)
    e.GET("/pixel/:x/:y/stats", getPixelStatsHandler)
    e.GET("/user/:user", getUserHandler)
    e.GET("/user/:user/stats", getUserStatsHandler)
    e.GET("/leaderboard", getLeaderboard)

    e.Logger.Fatal(e.Start(":8182"))
}