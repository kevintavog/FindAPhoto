package main

import (
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/kevintavog/findaphoto/common"
	clarifaifp "github.com/kevintavog/findaphoto/common/clarifai"
	clarifaiv2 "github.com/kevintavog/findaphoto/media-classifier/clarifai"

	"github.com/ian-kent/go-log/log"
	"github.com/twinj/uuid"
)

const imageMaxHeightDimension = 2048

// Clarifai allows up to 60 classification units per hour (one per image, one per second of video).
// We limit over a smaller period of time - so that restarting a session will probably work better
const throttleSecondDuration = 30
const maxPerThrottleDuration = 60.0 * float32(throttleSecondDuration)

type classifyThrottle struct {
	when  time.Time
	count int
}

var ClarifaiAPIKey string

var recentClassifications = make([]classifyThrottle, 0)

func classifyV2(mediaFile string) (string, error) {

	isImage := true
	if strings.ToLower(path.Ext(mediaFile)) == ".mp4" {
		isImage = false

		// For the V2 api, video uploads are limited to 10 MB
		fi, err := os.Stat(mediaFile)
		if err != nil {
			return "", err
		}
		if fi.Size() > 10*1024*1024 {
			log.Info("Skipping too large video: %s (%d)", mediaFile, fi.Size())
			return "", nil
		}
	}

	filename := mediaFile
	var generatedFilename string

	if strings.ToLower(path.Ext(mediaFile)) == ".jpg" {
		generatedFilename, err := generateSmallerImage(mediaFile)
		if err != nil {
			log.Error("Error generating smaller image for %s: (%s)", mediaFile, err)
		} else {
			filename = generatedFilename
		}
	}

	if len(generatedFilename) > 0 {
		defer os.Remove(generatedFilename)
	}

	log.Info("Classifying %s", mediaFile)
	client := clarifaiv2.NewClient(ClarifaiAPIKey)
	response, err := client.Predict(isImage, filename)
	if err != nil {
		return "", err
	}

	json := string(response[:])
	classifyComplete(mediaFile, json)
	return json, nil
}

func classifyComplete(mediaFile string, json string) {
	_, count, err := clarifaifp.TagsAndProbabilitiesFromJSON(json, 0)
	if err != nil {
		log.Fatalf("Failed getting tags and probabilities: %s (%s) [%s]", err, mediaFile, json)
		return
	}

	throttle := classifyThrottle{when: time.Now(), count: count}
	recentClassifications = append(recentClassifications, throttle)

	// Trim off those more than a few minutes old
	oldestAllowed := time.Now().Add(-1 * time.Second * throttleSecondDuration)
	for {
		if oldestAllowed.After(recentClassifications[0].when) {
			recentClassifications = recentClassifications[1:]
		} else {
			break
		}
	}

	// Calculate current throughput
	currentCount := 0
	for _, rc := range recentClassifications {
		currentCount += rc.count
	}

	if float32(currentCount) > maxPerThrottleDuration {
		pauseDuration := recentClassifications[0].when.Add(time.Second * throttleSecondDuration).Sub(time.Now())
		if pauseDuration.Seconds() > 6 {
			log.Warn("Exceeded threshold of %1.1f/second: %d in last %d seconds, pausing %s", maxPerThrottleDuration, currentCount, throttleSecondDuration, pauseDuration)
		}
		time.Sleep(pauseDuration)
	}
}

func generateSmallerImage(imageFilename string) (string, error) {
	tmpFilename := path.Join(os.TempDir(), "findAPhoto.media-classifier", "slides", uuid.NewV4().String()+".JPG")
	err := common.CreateDirectory(path.Dir(tmpFilename))
	if err != nil {
		return "", err
	}

	_, err = exec.Command(common.VipsThumbnailPath, "-d", "-s", "20000x"+strconv.Itoa(imageMaxHeightDimension), "-f", tmpFilename+"[optimize_coding,strip]", imageFilename).Output()
	if err != nil {
		return "", err
	}

	return tmpFilename, nil
}

func fakeClassifyV2(mediaFile string) (string, error) {
	isImage := true
	if strings.ToLower(path.Ext(mediaFile)) == ".mp4" {
		isImage = false
	}

	if isImage {
		response := `{"status":{"code":10000,"description":"Ok"},"outputs":[{"id":"c100422bd0334930a3e293c83b898895","status":{"code":10000,"description":"Ok"},"created_at":"2017-01-03T22:57:27Z","model":{"name":"general-v1.3","id":"aaa03c23b3724a16a56b629203edc62c","created_at":"2016-03-09T17:11:39Z","app_id":null,"output_info":{"message":"Show output_info with: GET /models/{model_id}/output_info","type":"concept"},"model_version":{"id":"aa9ca48295b37401f8af92ad1af0d91d","created_at":"2016-07-13T01:19:12Z","status":{"code":21100,"description":"Model trained successfully"}}},"input":{"id":"c100422bd0334930a3e293c83b898895","data":{"image":{"url":"https://s3.amazonaws.com/clarifai-api/img/prod/a82d78f28e544bad8a8b91387cb13c15/d60d0df8132d4ff3a843d6eeb46069f4.jpeg"}}},"data":{"concepts":[{"id":"ai_TZ3C79C6","name":"road","app_id":null,"value":0.98390055},{"id":"ai_GjVpxXrs","name":"street","app_id":null,"value":0.98112774},{"id":"ai_9Dcdh0PK","name":"house","app_id":null,"value":0.97597325},{"id":"ai_x3vjxJsW","name":"home","app_id":null,"value":0.9723978},{"id":"ai_rsX6XWc2","name":"building","app_id":null,"value":0.9714263},{"id":"ai_VRmbGVWh","name":"travel","app_id":null,"value":0.9657682},{"id":"ai_786Zr311","name":"no person","app_id":null,"value":0.9613738},{"id":"ai_WTrlNkqM","name":"vehicle","app_id":null,"value":0.9584619},{"id":"ai_FWCjC8jZ","name":"architecture","app_id":null,"value":0.9569262},{"id":"ai_MTvKbKJv","name":"landscape","app_id":null,"value":0.9524432},{"id":"ai_ZrPNDjxN","name":"daylight","app_id":null,"value":0.94442093},{"id":"ai_Zmhsv0Ch","name":"outdoors","app_id":null,"value":0.93465745},{"id":"ai_9pWqzvmM","name":"calamity","app_id":null,"value":0.9344537},{"id":"ai_LplXDDZ2","name":"car","app_id":null,"value":0.927792},{"id":"ai_WBQfVV0p","name":"city","app_id":null,"value":0.9193257},{"id":"ai_m8rrtkhG","name":"town","app_id":null,"value":0.9179833},{"id":"ai_0HffL2Dp","name":"environment","app_id":null,"value":0.916533},{"id":"ai_l8TKp2h5","name":"people","app_id":null,"value":0.91305476},{"id":"ai_j09mzT6j","name":"family","app_id":null,"value":0.9077202},{"id":"ai_Xxjc3MhT","name":"transportation system","app_id":null,"value":0.9031172}]}}]}`
		classifyComplete(mediaFile, response)
		return response, nil
	}

	response := `{"status":{"code":10000,"description":"Ok"},"outputs":[{"id":"f5856150ee0a4053926d4dfbc84f4a3c","status":{"code":10000,"description":"Ok"},"created_at":"2018-04-12T23:00:30.279453042Z","model":{"id":"aaa03c23b3724a16a56b629203edc62c","name":"general-v1.3","created_at":"2016-03-09T17:11:39.608845Z","app_id":"main","output_info":{"message":"Show output_info with: GET /models/{model_id}/output_info","type":"concept","type_ext":"concept"},"model_version":{"id":"aa9ca48295b37401f8af92ad1af0d91d","created_at":"2016-07-13T01:19:12.147644Z","status":{"code":21100,"description":"Model trained successfully"}},"display_name":"General"},"input":{"id":"f9a599c9b3dd46cd862302b087570ee7","data":{"video":{"url":"https://s3.amazonaws.com/clarifai-api/vid/prod/a82d78f28e544bad8a8b91387cb13c15/a089f240df8c409da205c10eff16d768","base64":"dHJ1ZQ=="}}},"data":{"frames":[{"frame_info":{"index":0,"time":0},"data":{"concepts":[{"id":"ai_l8TKp2h5","name":"people","value":0.98556685,"app_id":"main"},{"id":"ai_GjVpxXrs","name":"street","value":0.9852812,"app_id":"main"},{"id":"ai_TZ3C79C6","name":"road","value":0.9845048,"app_id":"main"},{"id":"ai_WTrlNkqM","name":"vehicle","value":0.96974087,"app_id":"main"},{"id":"ai_WBQfVV0p","name":"city","value":0.96771485,"app_id":"main"},{"id":"ai_f3fsJS61","name":"pavement","value":0.9530247,"app_id":"main"},{"id":"ai_4dB4Dh6F","name":"offense","value":0.9510988,"app_id":"main"},{"id":"ai_dxSG2s86","name":"man","value":0.9495779,"app_id":"main"},{"id":"ai_CpFBRWzD","name":"urban","value":0.9439218,"app_id":"main"},{"id":"ai_Xxjc3MhT","name":"transportation system","value":0.9411977,"app_id":"main"},{"id":"ai_VPmHr5bm","name":"adult","value":0.93535256,"app_id":"main"},{"id":"ai_vxpRR3nL","name":"police","value":0.92429155,"app_id":"main"},{"id":"ai_SVshtN54","name":"one","value":0.91990006,"app_id":"main"},{"id":"ai_tr0MBp64","name":"traffic","value":0.91876477,"app_id":"main"},{"id":"ai_ggQlMG6W","name":"industry","value":0.9046854,"app_id":"main"},{"id":"ai_6lhccv44","name":"business","value":0.89687645,"app_id":"main"},{"id":"ai_LplXDDZ2","name":"car","value":0.8871943,"app_id":"main"},{"id":"ai_86sS08Pw","name":"wear","value":0.8842746,"app_id":"main"},{"id":"ai_9pWqzvmM","name":"calamity","value":0.87620085,"app_id":"main"},{"id":"ai_w0Drrqn6","name":"accident","value":0.8684058,"app_id":"main"}]}},{"frame_info":{"index":1,"time":1000},"data":{"concepts":[{"id":"ai_l8TKp2h5","name":"people","value":0.98556685,"app_id":"main"},{"id":"ai_GjVpxXrs","name":"street","value":0.9852812,"app_id":"main"},{"id":"ai_TZ3C79C6","name":"road","value":0.9845048,"app_id":"main"},{"id":"ai_WTrlNkqM","name":"vehicle","value":0.96974087,"app_id":"main"},{"id":"ai_WBQfVV0p","name":"city","value":0.96771485,"app_id":"main"},{"id":"ai_f3fsJS61","name":"pavement","value":0.9530247,"app_id":"main"},{"id":"ai_4dB4Dh6F","name":"offense","value":0.9510988,"app_id":"main"},{"id":"ai_dxSG2s86","name":"man","value":0.9495779,"app_id":"main"},{"id":"ai_CpFBRWzD","name":"urban","value":0.9439218,"app_id":"main"},{"id":"ai_Xxjc3MhT","name":"transportation system","value":0.9411977,"app_id":"main"},{"id":"ai_VPmHr5bm","name":"adult","value":0.93535256,"app_id":"main"},{"id":"ai_vxpRR3nL","name":"police","value":0.92429155,"app_id":"main"},{"id":"ai_SVshtN54","name":"one","value":0.91990006,"app_id":"main"},{"id":"ai_tr0MBp64","name":"traffic","value":0.91876477,"app_id":"main"},{"id":"ai_ggQlMG6W","name":"industry","value":0.9046854,"app_id":"main"},{"id":"ai_6lhccv44","name":"business","value":0.89687645,"app_id":"main"},{"id":"ai_LplXDDZ2","name":"car","value":0.8871943,"app_id":"main"},{"id":"ai_86sS08Pw","name":"wear","value":0.8842746,"app_id":"main"},{"id":"ai_9pWqzvmM","name":"calamity","value":0.87620085,"app_id":"main"},{"id":"ai_w0Drrqn6","name":"accident","value":0.8684058,"app_id":"main"}]}}]}}]}`
	classifyComplete(mediaFile, response)
	return response, nil
}
