package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/crazyfacka/remoteclean/domain"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func deleteFile(conn *ssh.Client, f string) error {
	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	f = f[:len(f)-4]
	if err := session.Run("rm \"" + f + ".*\""); err != nil {
		return err
	}

	return nil
}

func getGigs(size string) float64 {
	var value float64

	scale := size[len(size)-1:]
	switch scale {
	case "G":
		value, _ = strconv.ParseFloat(size[:len(size)-1], 64)
	case "M":
		value, _ = strconv.ParseFloat(size[:len(size)-1], 64)
		value /= 1000
	case "K":
		value, _ = strconv.ParseFloat(size[:len(size)-1], 64)
		value /= (1000 * 1000)
	}

	return value
}

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
		if err := session.Run("find \"" + dir + "\" -type f -exec ls -1hRgp --full-time {} \\;"); err != nil {
			return nil, err
		}
	}

	re := regexp.MustCompile(`(?i).*\.(mp4|mkv)`)
	stapleTime := "2006-01-02 15:04:05"

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, line := range lines {
		if re.MatchString(line) {
			split := strings.Fields(line)
			gb := getGigs(split[3])
			t, _ := time.Parse(stapleTime, split[4]+" "+split[5])
			fpath := stringFromSlice(split[7:])
			items = append(items, domain.Item{
				Created:  t,
				FullPath: fpath,
				Size:     gb,
			})
		}
	}

	return domain.Items(items), nil
}

// GetFreeSpace returns the free space on the host
func GetFreeSpace(conn *ssh.Client, mount string) (float64, error) {
	var output bytes.Buffer

	session, err := conn.NewSession()
	if err != nil {
		return -1, err
	}

	session.Stdout = &output
	if err := session.Run("df -h \"" + mount + "\""); err != nil {
		return -1, err
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	gb := strings.Fields(lines[1])[3]
	value, err := strconv.ParseFloat(gb[:len(gb)-1], 64)
	if err != nil {
		return -1, err
	}

	return value, nil
}

// DeleteUntil deletes files until it reaches a certain amount of free space
func DeleteUntil(conn *ssh.Client, its domain.Items, current float64, until float64, dry bool) {
	if dry {
		log.Debug().Msg("Dryrun enabled")
	}

	amountDeleted := 0.0
	amountToDelete := until - current
	log.Info().Str("at_least", fmt.Sprintf("%.2fG", amountToDelete)).Msg("Erasing files")

	for _, i := range its {
		var err error
		if dry {
			err = nil
		} else {
			err = deleteFile(conn, i.FullPath)
		}

		if err == nil {
			amountDeleted += i.Size
			log.Info().Str("file", i.FullPath).Float64("amount", i.Size).Msg("Deleted")
		} else {
			log.Error().Str("file", i.FullPath).Err(err).Msg("Error deleting file")
		}

		if amountDeleted >= amountToDelete {
			break
		}
	}
}

// RefreshLibrary refreshes the player library
func RefreshLibrary(host string) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "VideoLibrary.Clean",
		"id":      1,
		"params": map[string]bool{
			"showdialogs": true,
		},
	})

	if err != nil {
		return err
	}

	log.Info().Msg("Refreshing player's library")
	resp, err := http.Post("http://"+host+":8080/jsonrpc", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
