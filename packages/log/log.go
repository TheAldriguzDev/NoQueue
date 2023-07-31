package log

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var Colors = Color{
	Black:        "\033[30m",
	Red:          "\033[31m",
	Green:        "\033[32m",
	Yellow:       "\033[33m",
	Blue:         "\033[34m",
	Magenta:      "\033[35m",
	Cyan:         "\033[36m",
	White:        "\033[37m",
	DarkGrey:     "\033[90m",
	LightGrey:    "\033[37m",
	Reset:        "\033[0m",
	Bold:         "\033[1m",
	Faint:        "\033[2m",
	Italic:       "\033[3m",
	Underline:    "\033[4m",
	BlinkSlow:    "\033[5m",
	BlinkRapid:   "\033[6m",
	ReverseVideo: "\033[7m",
}

type AvailableSites []string

var Sites = AvailableSites{"SkateDeluxe", "Footlocker", "Nike Order"}

// func PrintLogo() {
// 	fmt.Print(Colors.White + `
// ████████╗██╗███████╗███╗   ██╗███╗   ██╗███████╗     █████╗ ██╗ ██████╗
// ╚══██╔══╝██║██╔════╝████╗  ██║████╗  ██║██╔════╝    ██╔══██╗██║██╔═══██╗
//    ██║   ██║█████╗  ██╔██╗ ██║██╔██╗ ██║█████╗      ███████║██║██║   ██║
//    ██║   ██║██╔══╝  ██║╚██╗██║██║╚██╗██║██╔══╝      ██╔══██║██║██║   ██║
//    ██║   ██║███████╗██║ ╚████║██║ ╚████║███████╗    ██║  ██║██║╚██████╔╝
//    ╚═╝   ╚═╝╚══════╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝    ╚═╝  ╚═╝╚═╝ ╚═════╝
// ` + Colors.Reset)
// }

func PrintLogo() {
	fmt.Print(Colors.Yellow + `
██╗  ██╗██╗   ██╗██████╗ ███████╗███╗   ███╗███████╗██████╗ ██╗ █████╗ 
██║  ██║╚██╗ ██╔╝██╔══██╗██╔════╝████╗ ████║██╔════╝██╔══██╗██║██╔══██╗
███████║ ╚████╔╝ ██████╔╝█████╗  ██╔████╔██║█████╗  ██║  ██║██║███████║
██╔══██║  ╚██╔╝  ██╔═══╝ ██╔══╝  ██║╚██╔╝██║██╔══╝  ██║  ██║██║██╔══██║
██║  ██║   ██║   ██║     ███████╗██║ ╚═╝ ██║███████╗██████╔╝██║██║  ██║
╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚══════╝╚═╝     ╚═╝╚══════╝╚═════╝ ╚═╝╚═╝  ╚═╝												   
` + Colors.Reset)
}

func ClearConsole() {
	var clearCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		clearCmd = exec.Command("cmd", "/c", "cls")
	} else {
		clearCmd = exec.Command("clear")
	}
	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}

func GetDate() string {
	presentTime := time.Now()
	return presentTime.Format("01-02-2006")
}

func GetTime() string {
	presentTime := time.Now()
	return presentTime.Format("15:04:05")
}

func Log(message string, color string, loc string) {
	fmt.Printf(color+"\r[%s] [%s] "+message+" \n"+Colors.Reset, loc, GetTime())
}

func Error(msg string, loc string) {
	fmt.Printf(Colors.Red+"\r[%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime())
}

func Warn(msg string, loc string) {
	fmt.Printf(Colors.Yellow+"\r[%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime())
}

func Success(msg string, loc string) {
	fmt.Printf(Colors.Green+"\r[%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime())
}

func General(msg string, loc string) {
	fmt.Printf(Colors.DarkGrey+"\r[%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime())
}

func Info(msg string, loc string) {
	fmt.Printf(Colors.Magenta+"\r[%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime())
}

// bot logs

func LogTask(message string, color string, loc string, id string) {
	fmt.Printf(color+"\r[%s] [%s] [%s] "+message+" \n"+Colors.Reset, loc, GetTime(), id)
}

func ErrorTask(msg string, loc string, id string) {
	fmt.Printf(Colors.Red+"\r[%s] [%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime(), id)
}

func WarnTask(msg string, loc string, id string) {
	fmt.Printf(Colors.Yellow+"\r[%s] [%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime(), id)
}

func SuccessTask(msg string, loc string, id string) {
	fmt.Printf(Colors.Green+"\r[%s] [%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime(), id)
}

func GeneralTask(msg string, loc string, id string) {
	fmt.Printf(Colors.DarkGrey+"\r[%s] [%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime(), id)
}

func InfoTask(msg string, loc string, id string) {
	fmt.Printf(Colors.Magenta+"\r[%s] [%s] [%s] "+msg+" \n"+Colors.Reset, loc, GetTime(), id)
}

func PrettyId(num int) string {
	if num < 1 || num > 9999 {
		return "Numero non valido. Inserisci un numero tra 1 e 9999."
	}

	// Converti il numero in stringa
	numStr := strconv.Itoa(num)

	// Calcola il numero di zeri da aggiungere
	zeroesToAdd := 4 - len(numStr)

	// Costruisci la stringa con gli zeri davanti
	paddedNum := ""
	for i := 0; i < zeroesToAdd; i++ {
		paddedNum += "0"
	}
	paddedNum += numStr

	return paddedNum
}
