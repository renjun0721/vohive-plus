package audiotranscode

// Capabilities checks native symbols without opening a modem or recording audio.
// The Transcoder caches library handles, so repeated diagnostics do not reload them.
func (t *Transcoder) Capabilities() map[string]string {
	results := map[string]string{}
	for _, codec := range []string{"AMR", "AMR-WB"} {
		_, err := t.realtimeAPI(codec)
		if err != nil {
			results[codec] = err.Error()
		} else {
			results[codec] = ""
		}
	}
	t.lameOnce.Do(func() { t.lame, t.lameErr = loadNativeLame() })
	if t.lameErr != nil {
		results["MP3"] = t.lameErr.Error()
	} else {
		results["MP3"] = ""
	}
	return results
}
