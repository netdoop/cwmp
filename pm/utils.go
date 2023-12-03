package pm

import (
	"errors"
	"path"
	"regexp"
	"time"
)

// A20230627.2015+0800-2030+0800_000000.65740512A3200006L.xml
// A20230630.2345+0800-20230701.0000+0800_000000.65740512A3200006L.xml
// nrm_000000.65740512A3200006L.xml
func ParseFileName(filename string) (string, time.Time, time.Time, string, string, string, error) {
	filename = path.Base(filename)
	re := regexp.MustCompile(`A(\d{8})\.(\d{4})([+-]\d{4})-(\d{4})([+-]\d{4})_(\S+)\.(\S+)\.xml`)
	re2 := regexp.MustCompile(`A(\d{8})\.(\d{4})([+-]\d{4})-(\d{8})\.(\d{4})([+-]\d{4})_(\S+)\.(\S+)\.xml`)
	re3 := regexp.MustCompile(`nrm_(\S+)\.(\S+)\.xml`)
	re4 := regexp.MustCompile(`(\S+)\.(\S+)\.(\S+)_ConfigurationFile_(\d{14})`)
	re5 := regexp.MustCompile(`(\S+)\.(\S+)\.(\S+)_LogFile_(\d{14})`)

	now := time.Now()

	if matches := re.FindStringSubmatch(filename); len(matches) == 8 {
		date := matches[1]
		startTime := matches[2]
		startTimezone := matches[3]
		endTime := matches[4]
		endTimezone := matches[5]
		oui := matches[6]
		serialNumber := matches[7]

		dtFormat := "20060102 1504 -0700"
		start, err := time.Parse(dtFormat, date+" "+startTime+" "+startTimezone)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		end, err := time.Parse(dtFormat, date+" "+endTime+" "+endTimezone)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		return "PmFile", start.UTC(), end.UTC(), oui, "", serialNumber, nil
	}

	if matches := re2.FindStringSubmatch(filename); len(matches) == 9 {
		startDate := matches[1]
		startTime := matches[2]
		startTimezone := matches[3]

		endDate := matches[4]
		endTime := matches[5]
		endTimezone := matches[6]

		oui := matches[7]
		serialNumber := matches[8]

		dtFormat := "20060102 1504 -0700"
		start, err := time.Parse(dtFormat, startDate+" "+startTime+" "+startTimezone)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		end, err := time.Parse(dtFormat, endDate+" "+endTime+" "+endTimezone)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		return "PmFile", start.UTC(), end.UTC(), oui, "", serialNumber, nil
	}

	if matches := re3.FindStringSubmatch(filename); len(matches) == 3 {
		oui := matches[1]
		serialNumber := matches[2]
		return "NrmFile", now, now, oui, "", serialNumber, nil
	}

	if matches := re4.FindStringSubmatch(filename); len(matches) == 5 {
		oui := matches[1]
		productClass := matches[2]
		serialNumber := matches[3]
		timeString := matches[4]
		t, err := time.Parse("20060102030405", timeString)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		return "ConfigurationFile", t, t, oui, productClass, serialNumber, nil
	}

	if matches := re5.FindStringSubmatch(filename); len(matches) == 5 {
		oui := matches[1]
		productClass := matches[2]
		serialNumber := matches[3]
		timeString := matches[4]
		t, err := time.Parse("20060102030405", timeString)
		if err != nil {
			return "", now, now, "", "", "", errors.New("invalid filename format")
		}
		return "LogFile", t, t, oui, productClass, serialNumber, nil
	}
	return "", now, now, "", "", "", errors.New("invalid filename format")

}
