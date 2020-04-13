package handler

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/crazyfacka/remoteclean/domain"
	"golang.org/x/crypto/ssh"
)

func stringFromSlice(tokens []string) string {
	var sb strings.Builder

	for i, t := range tokens {
		sb.WriteString(t)
		if i < len(tokens)-1 {
			sb.WriteString(" ")
		}
	}

	return sb.String()
}

// GetContents lists all the contents from the remote host
func GetContents(conn *ssh.Client, dirs []string) (domain.Items, error) {
	var output bytes.Buffer
	var items []domain.Item

	for _, dir := range dirs {
		session, err := conn.NewSession()
		if err != nil {
			return nil, err
		}

		session.Stdout = &output
		if err := session.Run("find \"" + dir + "\" -type f -exec ls -1Rgp --full-time {} \\;"); err != nil {
			return nil, err
		}
	}

	re := regexp.MustCompile(`(?i).*\.(mp4|mkv)`)
	stapleTime := "2006-01-02 15:04:05"

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, line := range lines {
		if re.MatchString(line) {
			split := strings.Fields(line)
			t, _ := time.Parse(stapleTime, split[4]+" "+split[5])
			fpath := stringFromSlice(split[7:])
			items = append(items, domain.Item{
				Created:  t,
				FullPath: fpath,
			})
		}
	}

	return domain.Items(items), nil
}

// HasFreeSpace checks if there is free space on the host
func HasFreeSpace(conn *ssh.Client, mount string) (bool, error) {
	var output bytes.Buffer

	session, err := conn.NewSession()
	if err != nil {
		return true, err
	}

	session.Stdout = &output
	if err := session.Run("df -h \"" + mount + "\""); err != nil {
		return true, err
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	percentage := strings.Fields(lines[1])[4]
	value, err := strconv.Atoi(percentage[:len(percentage)-1])
	if err != nil {
		return true, err
	}

	if value > 50 {
		return false, nil
	}

	return true, nil
}
