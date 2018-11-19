package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/Nextdoor/conductor/core"
	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	testDataTypeFlag = flag.String("test_data_type", "full",
		"What kind of setup to do. [full] setup / [extend] the train / [create] a new train.")
)

func main() {
	flag.Parse()

	core.Preload()

	rand.Seed(time.Now().UTC().UnixNano())

	switch *testDataTypeFlag {
	case "full":
		full()
	case "extend":
		extend()
	case "create":
		create()
	}
	fmt.Println("Done.")
}

func full() {
	fmt.Println("Initializing random test data...")

	dataClient := data.NewClient()

	_, err := dataClient.ReadOrCreateUser("robot", "robot@example.com")
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(trains); i++ {
		fmt.Printf("Generating train %d\n", i)

		user, err := dataClient.ReadOrCreateUser(
			trains[i].User.Name, trains[i].User.Email)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Creating train...")
		train, err := dataClient.CreateTrain(
			trains[i].Branch, user, trains[i].Commits)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Completing delivery and verification phases...")
		if i == len(trains)-1 {
			err = dataClient.StartPhase(train.ActivePhases.Delivery)
			if err != nil {
				fmt.Println(err)
			}
			err = dataClient.CompletePhase(train.ActivePhases.Delivery)
			if err != nil {
				fmt.Println(err)
			}
			err = dataClient.StartPhase(train.ActivePhases.Verification)
			if err != nil {
				fmt.Println(err)
			}
			err = dataClient.CompletePhase(train.ActivePhases.Verification)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func extend() {
	fmt.Println("Extending train...")

	dataClient := data.NewClient()
	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		fmt.Println(err)
	}
	commit, err := dataClient.LatestCommitForTrain(latestTrain)
	if err != nil {
		fmt.Println(err)
	}

	sha := randomSHA()
	newCommits := []*types.Commit{
		{
			Message:     fmt.Sprintf("Extension Commit %d", commit.ID+1),
			AuthorName:  names[Variant1],
			AuthorEmail: emails[Variant1],
			URL:         fmt.Sprintf("https://github.com/%s", sha),
			SHA:         sha,
		},
	}
	messagingService := messaging.GetService()
	ticketService := ticket.GetService()
	core.ExtendTrain(dataClient, messagingService, latestTrain, newCommits, nil)
	core.StartTrain(dataClient, code.GetService(), messagingService, phase.GetService(), ticketService, latestTrain)
}

func create() {
	fmt.Println("Creating new train...")

	dataClient := data.NewClient()
	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		fmt.Println(err)
	}
	commit, err := dataClient.LatestCommitForTrain(latestTrain)
	if err != nil {
		fmt.Println(err)
	}
	shas := []string{randomSHA(), randomSHA(), randomSHA(), randomSHA(), randomSHA()}
	commits := []*types.Commit{
		{
			Message:     fmt.Sprintf("New Commit %d", commit.ID+1),
			AuthorName:  names[Variant1],
			AuthorEmail: emails[Variant1],
			URL:         fmt.Sprintf("https://github.com/%s", shas[0]),
			SHA:         shas[0],
		},
		{
			Message:     fmt.Sprintf("New Commit %d", commit.ID+2),
			AuthorName:  names[Variant2],
			AuthorEmail: emails[Variant2],
			URL:         fmt.Sprintf("https://github.com/%s", shas[1]),
			SHA:         shas[1],
		},
		{
			Message:     fmt.Sprintf("New Commit %d", commit.ID+3),
			AuthorName:  names[Variant3],
			AuthorEmail: emails[Variant3],
			URL:         fmt.Sprintf("https://github.com/%s", shas[2]),
			SHA:         shas[2],
		},
	}
	messagingService := messaging.GetService()
	ticketService := ticket.GetService()
	train := core.CreateTrain(dataClient, messagingService, "master", commits)
	core.StartTrain(dataClient, code.GetService(), messagingService, phase.GetService(), ticketService, train)
}

type DataKey int

const (
	Variant1 DataKey = iota
	Variant2
	Variant3
	Variant4
	Variant5
	Variant6
	Variant7
	Variant8
	Variant9
	Long
	Empty
	Multiline
	LongOneWord
	LongMultiword
	Spaces
	Cthulhu
)

var names = map[DataKey]string{
	Variant1:      "N",
	Variant2:      "Normal Name",
	Variant3:      "Thomas A. Anderson",
	LongOneWord:   "MyNameIsEndless...........................................................................................................................................................................................................................................",
	LongMultiword: "Person with a crazy long name, I mean it's just ridiculously long, why would anyone name their kid this?",
	Empty:         "",
}

var emails = map[DataKey]string{
	Variant1: "e",
	Variant2: "e.com",
	Variant3: "e@e.com",
	Spaces:   "e @ e . c o m",
	Long:     "personwithacrazylongemailImeanitsjustridiculouslylongwhywouldanyonemakethistheiremail@evenmyemaildomainnameissuperlong.com",
	Empty:    "",
}

func User(nameKey, emailKey DataKey) *types.User {
	return &types.User{
		Name:  names[nameKey],
		Email: emails[emailKey],
	}
}

var users = map[DataKey]*types.User{
	Variant1:    User(Variant1, Variant1),
	Variant2:    User(Variant2, Variant2),
	Variant3:    User(Variant3, Variant3),
	Long:        User(LongOneWord, Spaces),
	LongOneWord: User(LongOneWord, Long),
	Empty:       User(Empty, Empty),
}

var messages = map[DataKey]string{
	Variant1: "Commit message.",
	Variant2: "Initial commit.",
	Variant3: "Empty commit?",
	Variant4: "Tighten up the graphics on level 3.",
	Variant5: "If there is a bug please email me",
	Variant6: "Can someone please tell me what my code does?",
	Variant7: "Committed because I didn't do anything today.",
	Variant8: "Full refactor of the codebase.",
	Variant9: `Revert "Empty commit?"?`,
	Cthulhu:  `T͔̣̬͉̤̠̹̯̹̩̟̫̙́͟h͍͖̰̺̭̭̥̱̠̥̠̰͜ì̠̤̟̥͎̝̜̟̫̲̰̕̕͟ś̶̴̩̺̘̞̘̱͘͟ ̩̹̺̳̻͚̘̦̰̹͚͈̼́ͅć̜̜̯̣̤̳͚͇̟̲͎̹̭̼͓́͟͝ǫ͏̻̝̻̞̳̼̥̙̳̬͚͟͠͠m̢͞҉̵̯̙̥̞̹͇̟̩̞̬̺͇̰̗̠̖m͉̺͇̩͎̘̟͓̪̘͖͜͡i̴̫̳͕̗̝̤̞̙̠̝̠͇̦͖ͅţ̧̹̖̲̜͓̟̹̘̩̰̳̰̜ ͏̨̛͙̞̙͇͔̱͜s̵̹̯̠̫̞̳̞̟̱̙͇̯̹͉͈͘u̷̸͏̻̹̲͚̱̱̘͓̯̹̺ͅm̵̷̗͍͓̳̦̤̬̠͍͎͖̮̳̠̳̻̯m͏̷̪͍͖͔̻̝̬͎̝̹̳͓̗͈̘͙̲͜͢ơ̴̵̖̘͍̹̯̹̲͖͎̣̱̱͙̹̰͕͓͟͟n҉҉̨̝̟̜̲̙̰̺̣̭̩̪ś̷̴͙͔̤̯̤̯̮̦̻̠̫͖́ ̰͙͔̪̮̭͔̱̯̙̰̰̭͓̲̣͝͠͞Ć̶̴̘̖̱̙͓̩̤̺͕̰̼͚̖͟ͅͅͅͅt̶̫̫̝͎̕͟͞ͅh̢̖̫̖̠̩̗̣̱͇̥̲͇u̸̥̹͈̮̦̱̠͙̞͘͢ļ̬̠͉̩̠͚̮̫͚͉͘͠h̸̶͇͙̤̦̞̻̱̪̗̣̦̝͉̳̪͔̰ͅu̸̠͈̼͚̮͈͎͞҉̵̢͈̣̺̖͙̬̞͍̩̮͈̤̕̕ͅ`,
	Multiline: `Revert "Full refactor the entire codebase.".

The world is not ready`,
	LongOneWord: "thismessagemightgooffthescreen!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!, did it?",
	Empty:       "",
}

var currentSHA uint64 = 0

func Commit(messageKey, userKey DataKey) *types.Commit {
	currentSHA += 1

	user := users[userKey]
	sha := randomSHA()
	url := fmt.Sprintf("https://github.com/%s", sha)

	return &types.Commit{
		Message:     messages[messageKey],
		AuthorName:  user.Name,
		AuthorEmail: user.Email,
		URL:         url,
		SHA:         sha,
	}
}

var commits = []*types.Commit{
	Commit(Variant1, Variant1),
	Commit(Variant2, Variant2),
	Commit(Variant3, Variant3),
	Commit(Variant4, Variant1),
	Commit(Variant5, Variant2),
	Commit(Variant6, Variant3),
	Commit(Variant7, Variant1),
	Commit(Variant8, Variant2),
	Commit(Variant9, Variant3),
	Commit(Cthulhu, LongOneWord),
	Commit(Multiline, Long),
	Commit(LongOneWord, LongOneWord),
	Commit(Empty, Empty),
}

var branches = map[DataKey]string{
	Variant1: "master",
	Variant2: "branch",
	Variant3: "branch_1",
	Variant4: "branch_1a",
	Variant5: "b",
	Long:     "superlongbranchnameitjustkeepsgoingandgoingandgoingandgoingandgoingandgoingandgoinganditsalloneword",
	Empty:    "",
}

type TrainInfo struct {
	Commits []*types.Commit
	Branch  string
	User    *types.User
}

func Train(commitIndicies []int, branchKey DataKey, userKey DataKey) TrainInfo {
	trainCommits := make([]*types.Commit, len(commitIndicies))
	for i, index := range commitIndicies {
		trainCommits[i] = commits[index]
	}
	return TrainInfo{
		Commits: trainCommits,
		Branch:  branches[branchKey],
		User:    users[userKey],
	}
}

var trains = []TrainInfo{
	Train([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, Variant1, Variant1),
	Train([]int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, Variant2, Variant2),
	Train([]int{0, 12}, Variant3, Long),
	Train([]int{0, 1, 2, 3, 4}, Variant4, LongOneWord),
	Train([]int{8, 9, 10, 11, 12}, Variant5, Empty),
	Train([]int{8, 12}, Long, Variant1),
	Train([]int{0, 3, 5, 7, 9}, Variant1, Variant1),
}

func randomSHA() string {
	var runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, 40)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}
