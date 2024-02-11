package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "github.com/gordonklaus/portaudio"
    "encoding/binary"
)

var (
    filename string
)

func init() {
    flag.StringVar(&filename, "f", "output.wav", "output filename")
    flag.Parse()
}

func main() {
    // Create the output WAV file
    file, err := os.Create(filename)
    if err != nil {
        fmt.Println("Error creating output file:", err)
        return
    }
    defer file.Close()

    // Write WAV header
    err = writeWavHeader(file)
    if err != nil {
        fmt.Println("Error writing WAV header:", err)
        return
    }

    // Initialize PortAudio
    err = portaudio.Initialize()
    if err != nil {
        fmt.Println("PortAudio initialization error:", err)
        return
    }
    defer portaudio.Terminate()

    // Open a stream to record audio
    inputChannels := 1
    outputChannels := 0
    sampleRate := 44100
    framesPerBuffer := 1024
    stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), framesPerBuffer, recordCallback)
    if err != nil {
        fmt.Println("PortAudio stream error:", err)
        return
    }
    defer stream.Close()

    // Start the stream
    err = stream.Start()
    if err != nil {
        fmt.Println("PortAudio stream start error:", err)
        return
    }
    defer stream.Stop()

    // Wait for SIGINT or SIGTERM to stop recording
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    fmt.Println("Recording. Press Ctrl+C to stop...")
    <-sigCh
}

// recordCallback is called whenever PortAudio needs more audio data
func recordCallback(in []float32) {
    // Write audio data to the output file
    err := writeFloat32SamplesToFile(filename, in)
    if err != nil {
        fmt.Println("Error writing audio data:", err)
    }
}

// writeWavHeader writes the WAV file header
func writeWavHeader(file *os.File) error {
    // WAV header constants
    const (
        riffHeader      = "RIFF"
        waveHeader      = "WAVE"
        fmtHeader       = "fmt "
        dataHeader      = "data"
        pcmFormat       = 1
        bitsPerSample   = 16
        byteRate        = 44100 * 1 * bitsPerSample / 8
        blockAlign      = 1 * bitsPerSample / 8
    )

    // Write RIFF header
    _, err := file.WriteString(riffHeader)
    if err != nil {
        return err
    }
    // Write file size (placeholder)
    err = binary.Write(file, binary.LittleEndian, uint32(0))
    if err != nil {
        return err
    }
    // Write WAVE header
    _, err = file.WriteString(waveHeader)
    if err != nil {
        return err
    }
    // Write fmt header
    _, err = file.WriteString(fmtHeader)
    if err != nil {
        return err
    }
    // Write fmt chunk size
    err = binary.Write(file, binary.LittleEndian, uint32(16))
    if err != nil {
        return err
    }
    // Write audio format (PCM)
    err = binary.Write(file, binary.LittleEndian, uint16(pcmFormat))
    if err != nil {
        return err
    }
    // Write number of channels
    err = binary.Write(file, binary.LittleEndian, uint16(1))
    if err != nil {
        return err
    }
    // Write sample rate
    err = binary.Write(file, binary.LittleEndian, uint32(44100))
    if err != nil {
        return err
    }
    // Write byte rate
    err = binary.Write(file, binary.LittleEndian, uint32(byteRate))
    if err != nil {
        return err
    }
    // Write block align
    err = binary.Write(file, binary.LittleEndian, uint16(blockAlign))
    if err != nil {
        return err
    }
    // Write bits per sample
    err = binary.Write(file, binary.LittleEndian, uint16(bitsPerSample))
    if err != nil {
        return err
    }
    // Write data header
    _, err = file.WriteString(dataHeader)
    if err != nil {
        return err
    }
    // Write data size (placeholder)
    err = binary.Write(file, binary.LittleEndian, uint32(0))
    if err != nil {
        return err
    }
    return nil
}

// writeFloat32SamplesToFile writes float32 audio samples to a file
func writeFloat32SamplesToFile(filename string, data []float32) error {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)		//data should be written at the end of the file, os.O.WRONLY: this flag indicates that the file should be opended in write only mode
	
	if err != nil {
        return err
    }
    defer file.Close()

    for _, sample := range data {
        // Convert float32 sample to int16 for 16-bit WAV format
        intSample := int16(sample * (1<<15 - 1))
		//1 is shifted 15 times to the left and sub it by 1 gives the signed 16-bit integer[-32767, 32767]
        err := binary.Write(file, binary.LittleEndian, intSample)
        if err != nil {
            return err
        }
    }
    return nil
}
