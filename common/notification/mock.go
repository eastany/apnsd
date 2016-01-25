package notification

const testDeviceToken = "e93b7686988b4b5fd334298e60e73d90035f6d12628a80b4029bde0dec514df9"

func mockPayload() (payload *Payload) {
	payload = NewPayload()
	payload.Alert = "You have mail!"
	payload.Badge = 42
	payload.Sound = "bingbong.aiff"
	return
}

func mockAlertDictionary() (dict *AlertDictionary) {
	args := make([]string, 1)
	args[0] = "localized args"
	dict = NewAlertDictionary()
	dict.Title = "Title"
	dict.TitleLocKey = "localized  key"
	dict.TitleLocArgs = args
	dict.Body = "Complex Message"
	dict.ActionLocKey = "Play a Game!"
	dict.LocKey = "localized key"
	dict.LocArgs = args
	dict.LaunchImage = "image.jpg"
	return
}

func MockNotification() (data []byte) {
	payload := mockPayload()
	pn := NewNotification()
	pn.DeviceToken = testDeviceToken
	pn.AddPayload(payload)
	data, _ = pn.ToBytes()
	return
}
