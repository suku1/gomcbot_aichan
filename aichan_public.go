package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
	"regexp"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	crand "crypto/rand"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	bot "github.com/Tnze/gomcbot"
	auth "github.com/Tnze/gomcbot/authenticate"
)

var back *regexp.Regexp = regexp.MustCompile(`^(バックアップしています|Complete.)$`)
var omikuji *regexp.Regexp = regexp.MustCompile(`^(/omikuji|/o)$`)
var kaigyo *strings.Replacer = strings.NewReplacer("\n", "")
var aichan *regexp.Regexp = regexp.MustCompile(`(?i)^(aichan|あいちゃん|ｉちゃん)$`)
var reload *regexp.Regexp = regexp.MustCompile(`(?i)^(reload|レォあd)$`)
var aichan_reload *regexp.Regexp = regexp.MustCompile(`(?i)^(ｉちゃんレォあd)$`)

type Data struct {
    Id string `json:"id"`
    Password string `json:"password"`
    Server string `json:"server"`
    Port int `json:"port"`
}

type Userlist struct {
	telegram_list string
	ids []int64
	names []string
}

var UL Userlist

func analyze_chat(game *bot.Game, txt string) (err error) {
	tmp1 := strings.Split(strings.Split(txt, ">")[0], "<")
	if len(tmp1) < 2 {
		return fmt.Errorf("txt is not message")
	} 
	name := tmp1[1]
	msg := strings.Split(txt, " ")
	if len(msg) < 2 {
		return fmt.Errorf("txt is not message")
	} 
	msg = msg[1:]
	if name == "Bot_name" {
		if len(msg) < 2 {
			return fmt.Errorf("txt is not message")
		} 
		name = msg[0]
		l := len(name)
		name = name[1:l-1]
		msg = msg[1:]
	}
	if len(msg) >= 2 && aichan.MatchString(msg[0]){
		name_called(game, name, msg[1:])
	} else {
		chat_func(game, name, msg)
	}

	return nil
}

func chat_func(game *bot.Game, name string, msg []string) {
	if aichan_reload.MatchString(msg[0]) && name == "admin"{
		err := get_userlist()
		if err == nil {
			game.Chat("ユーザーリストを更新しました")
		} else {
			game.Chat("エラーです")
		}
	}
}

func name_called(game *bot.Game, name string, msg []string) {
	if reload.MatchString(msg[0]) && name == "admin"{
		err := get_userlist()
		if err == nil {
			game.Chat("ユーザーリストを更新しました")
		} else {
			game.Chat("エラーです")
		}
	}
}

func sendmsg(telegram *tgbotapi.BotAPI, txt string) {
	if back.MatchString(txt){ 
		return 
	}
	for _, id := range UL.ids {
		msg := tgbotapi.NewMessage(id, txt)
		//fmt.Println(id, txt)
		telegram.Send(msg)
	}
}

//*
func receivemsg(game *bot.Game, telegram *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		var txt string = update.Message.Text
		txt = kaigyo.Replace(txt)
		log.Printf("%d", update.Message.Chat.ID)
		log.Printf("[%s] %s", update.Message.From.UserName, txt)
		var check bool = true
		for i, id := range UL.ids {
			if update.Message.Chat.ID == id {
				check = false
				if omikuji.MatchString(txt) {
					game.Chat("/omikuji draw " + UL.names[i])
				} else {
					game.Chat("[" + UL.names[i] + "] " + txt)
				}
			}
		}
		if check == true {
			num, _ := rand_num()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, num)
			telegram.Send(msg)
			fmt.Println("Random number: " + num)
		}
	}
}
//*/
func rand_num() (res string, err error){
	num := strconv.Itoa(rand.Intn(1000000))
	digit := len(num)
	res = ""
	for i := 0; i < (6-digit); i++ {
		res += "0"
	}
	res += num
	return res, nil
}

func get_userlist() (err error){
	ls, err := os.Open(UL.telegram_list)
    if err != nil{
        return fmt.Errorf("Cannot open file")
    }
	defer ls.Close()

	UL.ids = []int64{}
	UL.names = []string{}
	reader := bufio.NewReaderSize(ls, 4096)
    for {
        line, _, err := reader.ReadLine()
        if err != nil{
            break
        }
		var s []string = strings.Split(string(line), " ")
		var tmp int64
		tmp, _ = strconv.ParseInt(s[0], 10, 64)
		UL.ids = append(UL.ids, tmp)
		UL.names = append(UL.names, s[1])
	}
	return nil
}

func main() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())

	UL.telegram_list = "telegram_list.txt"
	var telegram_apikey string = "API key"
	datafile := "data.json"
	f, err := ioutil.ReadFile(datafile)
    if err != nil {
		panic("error")
    }
	data := new(Data)
	err = json.Unmarshal(f, data)
	if err != nil {
        panic("error")
	}
	resp, err := auth.Authenticate(data.Id, data.Password)

	
	telegram, err := tgbotapi.NewBotAPI(telegram_apikey)

	//fmt.Printf("(%%#v) %#v\n", resp)
	
	if err != nil {
		panic(err)
	}
	Auth := resp.ToAuth()

	game, err := Auth.JoinServer(data.Server, data.Port)
	if err != nil {
		panic(err)
	}

	events := game.GetEvents()
	go game.HandleGame()
	
	if err != nil {
		log.Panic(err)
	}

	get_userlist()

	log.Printf("Authorized on account %s", telegram.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1
	updates, err := telegram.GetUpdatesChan(u)
	go receivemsg(game, telegram, updates)

	for e := range events {
		switch e.(type) {
		case bot.PlayerSpawnEvent:
			fmt.Println("ログイン成功")
		case bot.ChatMessageEvent:
			fmt.Println(e.(bot.ChatMessageEvent).Msg)
			var txt string
			for _, v := range e.(bot.ChatMessageEvent).Msg.Extra { txt += v.Text }
			go sendmsg(telegram, txt)
			go analyze_chat(game, txt)
		}

	}
}