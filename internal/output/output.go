package output

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	colorReset   = "\033[0m"
	colorGreen   = "\033[92m"
	colorYellow  = "\033[93m"
	colorRed     = "\033[91m"
)

func Green(text string) string  { return colorGreen + text + colorReset }
func Yellow(text string) string { return colorYellow + text + colorReset }
func Red(text string) string    { return colorRed + text + colorReset }

func Clear() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func PrintInfo(color func(string) string, v1, v2 string, v3 any) {
	align := 35
	if _, ok := v3.(int); ok {
		align = -35
	}
	fmt.Printf("%s\n", color(fmt.Sprintf("%-10s %-20s %-20s %*v %-5s", "*****", v1, v2, align, v3, "*****")))
}

func PrintError(messages []string, level string) {
	color := Red
	if level == "INFO" {
		color = Yellow
	}

	maxLen := 0
	for _, msg := range messages {
		if l := len(msg); l > maxLen {
			maxLen = l
		}
	}
	if maxLen+6 > 98 {
		maxLen = 92
	} else {
		maxLen += 6
	}

	boxWidth := maxLen + 4
	msgWidth := maxLen + 2
	blank := color("║" + strings.Repeat(" ", boxWidth) + "║")

	fmt.Println(color("\n╔" + strings.Repeat("=", boxWidth) + "╗"))
	fmt.Println(blank)
	fmt.Println(color("║"), color(fmt.Sprintf("%-*s", msgWidth, level+"!")), color("║"))
	fmt.Println(blank)

	for _, msg := range messages {
		if len(msg) > 98 {
			for i := 0; i < len(msg); i += 98 {
				end := i + 98
				if end > len(msg) {
					end = len(msg)
				}
				chunk := msg[i:end]
				fmt.Println(color("║"), color(center(chunk, msgWidth)), color("║"))
			}
		} else {
			fmt.Println(color("║"), color(center(msg, msgWidth)), color("║"))
		}
	}

	fmt.Println(blank)
	fmt.Println(blank)
	fmt.Println(color("╚" + strings.Repeat("=", boxWidth) + "╝\n"))
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}