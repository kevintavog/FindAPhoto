package main

import (
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/kevintavog/findaphoto/media-classifier/clarifai"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/common/clarifai"

	"github.com/ian-kent/go-log/log"
	"github.com/twinj/uuid"
	"github.com/zquestz/clarifai-go"
)

const imageMaxHeightDimension = 2048

// Clarifai allows up to 1000 classification units per hour (one per image, one per second of video).
// We limit over a smaller period of time - restarting a session will probably work better
const throttleSecondDuration = 30
const maxPerThrottleDuration = 1000.0 / 60.0 / 60.0 * float32(throttleSecondDuration)

type classifyThrottle struct {
	when  time.Time
	count int
}

var ClientId string
var ClientSecret string

var recentClassifications = make([]classifyThrottle, 0)

func checkClarifai() error {
	client := clarifai.NewClient(ClientId, ClientSecret)
	info, err := client.Info()
	if err != nil {
		return err
	}

	log.Info("clarifai.com info response: '%s'", info.StatusCode)
	return nil
}

func classifyV2(mediaFile string) (string, error) {

	if strings.ToLower(path.Ext(mediaFile)) == ".mp4" {
		log.Info("Skipping video: %s", mediaFile)
		return "", nil
	}

	filename := mediaFile
	var generatedFilename string = ""

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

	client := clarifaiv2.NewClient(ClientId, ClientSecret)
	response, err := client.Predict(filename)
	if err != nil {
		if client.Throttled {
			log.Fatalf("Throttled by Clarifai, exiting: %s", err)
		}
		return "", err
	}

	json := string(response[:])

	// There doesn't seem to be hourly throttling for V2 - so don't include it in throttling
	// (this does verify the JSON is parseable)
	classifyComplete(json, false)

	return json, nil
}

func classifyV1(mediaFile string) (string, error) {
	filename := mediaFile
	var generatedFilename string = ""

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

	client := clarifai.NewClient(ClientId, ClientSecret)
	response, err := client.Tag(clarifai.TagRequest{Files: []string{filename}})
	if err != nil {
		if client.Throttled {
			//			log.Fatalf("Throttled by Clarifai, exiting: %s", err)
		}
		return "", err
	}

	json := string(response[:])
	classifyComplete(json, true)

	return json, nil
}

func classifyComplete(json string, includeInThrottle bool) {
	_, count, err := clarifaifp.TagsAndProbabilitiesFromJson(json, 0)
	if err != nil {
		log.Fatalf("Failed getting tags and probabilities: %s (%s)", err, json)
		return
	}

	if !includeInThrottle {
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

func fakeClassify(mediaFile string) (string, error) {
	// Image
	response := `{"status_code": "OK", "status_msg": "All images in request have completed successfully. ", "meta": {"tag": {"timestamp": 1483038928.675021, "model": "general-v1.3", "config": null}}, "results": [{"docid": 329736847191055169825307895473890431497, "status_code": "OK", "status_msg": "OK", "local_id": "", "result": {"tag": {"classes": ["car", "vehicle", "drive", "transportation system", "road", "rally", "wheel", "driver", "traffic", "competition", "accident", "automotive", "truck", "hurry", "fast", "police", "race", "exhibition", "street", "wheel"], "concept_ids": ["ai_LplXDDZ2", "ai_WTrlNkqM", "ai_hsG8bzf6", "ai_Xxjc3MhT", "ai_TZ3C79C6", "ai_cFjHhsk5", "ai_gRZHdRD7", "ai_1Q53NvNw", "ai_tr0MBp64", "ai_7WNVdPhm", "ai_w0Drrqn6", "ai_K2d5Ps8r", "ai_DPX3d5qs", "ai_9WFnldtB", "ai_dSCKh8xv", "ai_vxpRR3nL", "ai_5WW7fH4K", "ai_9c0Hmcx0", "ai_GjVpxXrs", "ai_TW9GpHcz"], "probs": [0.9990969300270081, 0.9989654421806335, 0.9889678955078125, 0.9856127500534058, 0.9805673360824585, 0.9750142097473145, 0.9746156930923462, 0.9396674633026123, 0.9392398595809937, 0.9322007298469543, 0.9293274879455566, 0.9291046261787415, 0.9124788045883179, 0.9106875658035278, 0.9080287218093872, 0.9001490473747253, 0.8900867700576782, 0.8861268758773804, 0.8802436590194702, 0.862737774848938]}}, "docid_str": "f81101bc286119252a8864bc35149209"}]}`

	// Video
	//	response := `{"status_code": "OK", "status_msg": "All images in request have completed successfully. ", "meta": {"tag": {"timestamp": 1483039091.908673, "model": "general-v1.3", "config": null}}, "results": [{"docid": 168140403947542947042016930178390303122, "status_code": "OK", "status_msg": "OK", "local_id": "", "result": {"tag": {"timestamps": [0.0, 1.0, 2.0], "classes": [["blur", "music", "performance", "light", "motion", "festival", "people", "evening", "travel", "dancing", "city", "concert", "illuminated", "stage", "crowd", "celebration", "party", "musician", "building", "water"], ["music", "performance", "light", "festival", "people", "concert", "evening", "stage", "city", "party", "motion", "blur", "celebration", "travel", "singer", "musician", "dancing", "Christmas", "building", "crowd"], ["music", "performance", "light", "festival", "people", "concert", "evening", "stage", "city", "party", "motion", "blur", "celebration", "travel", "singer", "musician", "dancing", "Christmas", "building", "crowd"]], "concept_ids": [["ai_l4WckcJN", "ai_k76BrtPJ", "ai_325FlCBf", "ai_n9vjC1jB", "ai_FwtMR9mk", "ai_13NdwKqz", "ai_l8TKp2h5", "ai_PJQHT1jg", "ai_VRmbGVWh", "ai_vsnjzvPC", "ai_WBQfVV0p", "ai_sMbNHMwW", "ai_ccZ7tNd3", "ai_213hCgg9", "ai_dJ15S9s6", "ai_wmbvr5TG", "ai_b01mhdxB", "ai_S81BZrtF", "ai_rsX6XWc2", "ai_BlL0wSQh"], ["ai_k76BrtPJ", "ai_325FlCBf", "ai_n9vjC1jB", "ai_13NdwKqz", "ai_l8TKp2h5", "ai_sMbNHMwW", "ai_PJQHT1jg", "ai_213hCgg9", "ai_WBQfVV0p", "ai_b01mhdxB", "ai_FwtMR9mk", "ai_l4WckcJN", "ai_wmbvr5TG", "ai_VRmbGVWh", "ai_32STZ9b3", "ai_S81BZrtF", "ai_vsnjzvPC", "ai_mKzmkKDG", "ai_rsX6XWc2", "ai_dJ15S9s6"], ["ai_k76BrtPJ", "ai_325FlCBf", "ai_n9vjC1jB", "ai_13NdwKqz", "ai_l8TKp2h5", "ai_sMbNHMwW", "ai_PJQHT1jg", "ai_213hCgg9", "ai_WBQfVV0p", "ai_b01mhdxB", "ai_FwtMR9mk", "ai_l4WckcJN", "ai_wmbvr5TG", "ai_VRmbGVWh", "ai_32STZ9b3", "ai_S81BZrtF", "ai_vsnjzvPC", "ai_mKzmkKDG", "ai_rsX6XWc2", "ai_dJ15S9s6"]], "probs": [[0.993641197681427, 0.9930315017700195, 0.9899797439575195, 0.9889219403266907, 0.9876085519790649, 0.985458493232727, 0.9800918102264404, 0.979888916015625, 0.9709082245826721, 0.9684736728668213, 0.9661855697631836, 0.9597376585006714, 0.9576658010482788, 0.9445781707763672, 0.9420493841171265, 0.9342305660247803, 0.927289605140686, 0.9232202768325806, 0.92054283618927, 0.9098004102706909], [0.9960581064224243, 0.9899797439575195, 0.9856935739517212, 0.985458493232727, 0.9767289161682129, 0.9756318926811218, 0.9751339554786682, 0.9685540795326233, 0.9661855697631836, 0.9611333608627319, 0.9590389728546143, 0.9586986303329468, 0.9528767466545105, 0.9412574172019958, 0.9411396980285645, 0.9382811784744263, 0.9355343580245972, 0.9270551204681396, 0.92054283618927, 0.920181393623352], [0.9960581064224243, 0.9899797439575195, 0.9856935739517212, 0.985458493232727, 0.9767289161682129, 0.9756318926811218, 0.9751339554786682, 0.9685540795326233, 0.9661855697631836, 0.9611333608627319, 0.9590389728546143, 0.9586986303329468, 0.9528767466545105, 0.9412574172019958, 0.9411396980285645, 0.9382811784744263, 0.9355343580245972, 0.9270551204681396, 0.92054283618927, 0.920181393623352]]}}, "docid_str": "7e7ea9f432521664edb8e67c2b382192"}]}`

	classifyComplete(response, true)
	return response, nil
}

func fakeClassifyV2(mediaFile string) (string, error) {
	response := `{"status":{"code":10000,"description":"Ok"},"outputs":[{"id":"c100422bd0334930a3e293c83b898895","status":{"code":10000,"description":"Ok"},"created_at":"2017-01-03T22:57:27Z","model":{"name":"general-v1.3","id":"aaa03c23b3724a16a56b629203edc62c","created_at":"2016-03-09T17:11:39Z","app_id":null,"output_info":{"message":"Show output_info with: GET /models/{model_id}/output_info","type":"concept"},"model_version":{"id":"aa9ca48295b37401f8af92ad1af0d91d","created_at":"2016-07-13T01:19:12Z","status":{"code":21100,"description":"Model trained successfully"}}},"input":{"id":"c100422bd0334930a3e293c83b898895","data":{"image":{"url":"https://s3.amazonaws.com/clarifai-api/img/prod/a82d78f28e544bad8a8b91387cb13c15/d60d0df8132d4ff3a843d6eeb46069f4.jpeg"}}},"data":{"concepts":[{"id":"ai_TZ3C79C6","name":"road","app_id":null,"value":0.98390055},{"id":"ai_GjVpxXrs","name":"street","app_id":null,"value":0.98112774},{"id":"ai_9Dcdh0PK","name":"house","app_id":null,"value":0.97597325},{"id":"ai_x3vjxJsW","name":"home","app_id":null,"value":0.9723978},{"id":"ai_rsX6XWc2","name":"building","app_id":null,"value":0.9714263},{"id":"ai_VRmbGVWh","name":"travel","app_id":null,"value":0.9657682},{"id":"ai_786Zr311","name":"no person","app_id":null,"value":0.9613738},{"id":"ai_WTrlNkqM","name":"vehicle","app_id":null,"value":0.9584619},{"id":"ai_FWCjC8jZ","name":"architecture","app_id":null,"value":0.9569262},{"id":"ai_MTvKbKJv","name":"landscape","app_id":null,"value":0.9524432},{"id":"ai_ZrPNDjxN","name":"daylight","app_id":null,"value":0.94442093},{"id":"ai_Zmhsv0Ch","name":"outdoors","app_id":null,"value":0.93465745},{"id":"ai_9pWqzvmM","name":"calamity","app_id":null,"value":0.9344537},{"id":"ai_LplXDDZ2","name":"car","app_id":null,"value":0.927792},{"id":"ai_WBQfVV0p","name":"city","app_id":null,"value":0.9193257},{"id":"ai_m8rrtkhG","name":"town","app_id":null,"value":0.9179833},{"id":"ai_0HffL2Dp","name":"environment","app_id":null,"value":0.916533},{"id":"ai_l8TKp2h5","name":"people","app_id":null,"value":0.91305476},{"id":"ai_j09mzT6j","name":"family","app_id":null,"value":0.9077202},{"id":"ai_Xxjc3MhT","name":"transportation system","app_id":null,"value":0.9031172}]}}]}`
	classifyComplete(response, false)
	return response, nil
}
