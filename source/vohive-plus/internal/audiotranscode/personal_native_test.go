package audiotranscode

import (
	"os"
	"testing"
)

// Opt-in integration check uses synthesized silence, never a modem or a call.
func TestPersonalNativeCodecsWithSyntheticAudio(t *testing.T) {
	if os.Getenv("VOHIVE_PLUS_CODEC_TEST") != "1" {
		t.Skip("native runtime integration check is opt-in")
	}
	transcoder := New()
	for _, name := range []string{"AMR", "AMR-WB"} {
		t.Run(name, func(t *testing.T) {
			api, err := transcoder.realtimeAPI(name)
			if err != nil {
				t.Fatal(err)
			}
			configuration, err := realtimeConfig(name, api)
			if err != nil {
				t.Fatal(err)
			}
			codec, err := newRealtimeCodec(configuration, configuration.maxMode)
			if err != nil {
				t.Fatal(err)
			}
			defer codec.Close()
			codec.octetAligned = true
			encoded, err := codec.Encode(make([]int16, configuration.samplesPerFrame))
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := codec.Decode(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if len(decoded) != configuration.samplesPerFrame {
				t.Fatalf("decoded samples=%d", len(decoded))
			}
		})
	}
	lame, err := loadNativeLame()
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir() + "/silence.mp3"
	if err := encodeMP3File(t.Context(), mp3EncodeRequest{api: lame, outputPath: output, pcm: make([]int16, 8000), sampleRate: 8000}); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatal("MP3 encoding did not produce an audio file")
	}
}
