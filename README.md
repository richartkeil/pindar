# Pindar

A command-line tool for transcribing audio files using OpenAI's Whisper API.

## Features

- **Audio Transcription**: Transcribe audio files in various formats (unknown formats are automatically converted using ffmpeg)
- **Multiple Output Formats**: Support for text, SRT, VTT, and verbose JSON output
- **Flexible Configuration**: Set OpenAI API key via command line, environment variable, or persistent config
- **Custom Output Control**: Specify output directory and file extensions
- **Language Detection**: Automatic language detection or manual specification
- **Prompt Support**: Guide transcription with custom prompts

## Installation

### Prerequisites

- Go 1.19 or later
- **ffmpeg** (required for automatic audio format conversion)
  - macOS: `brew install ffmpeg`
  - Ubuntu/Debian: `sudo apt install ffmpeg`
  - Windows: Download from [ffmpeg.org](https://ffmpeg.org/download.html)

### Install from Source

```bash
git clone https://github.com/richartkeil/pindar.git
cd pindar
go build
```

### Install Globally

To install Pindar globally and add it to your PATH:

```bash
# Install to $GOPATH/bin (make sure $GOPATH/bin is in your PATH)
go install github.com/richartkeil/pindar@latest

# Or build and copy to a directory in your PATH
git clone https://github.com/richartkeil/pindar.git
cd pindar
go build
sudo cp pindar /usr/local/bin/
```

Make sure your PATH includes the installation directory:
- For `go install`: Ensure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH
- For manual installation: `/usr/local/bin` should already be in your PATH

## Usage

```bash
pindar audio.mp3
```

### Command Line Options

```bash
pindar [OPTIONS] <audio-file>

Options:
  --model string        OpenAI model to use (default: whisper-1)
  --language string     Language of the audio file (optional, auto-detected if not specified)
  --prompt string       Optional text to guide the model's style
  --format string       Output format: text, srt, verbose_json, or vtt (default: text)
  --output-dir, -o string    Directory to save output (default: current directory)
  --output-ext string   Custom extension for output file
  --api-key string      OpenAI API key (can also be set via OPENAI_API_KEY environment variable)
  --temperature float   Sampling temperature between 0 and 1 (default: 0)
```

### Examples

```bash
# Basic transcription
pindar interview.mp3

# Specify language and output format
pindar --language en --format srt meeting.wav

# Custom output directory and extension
pindar --output-dir ./transcripts --output-ext .transcript audio.m4a

# Use custom prompt for better context
pindar --prompt "This is a technical discussion about software development" podcast.mp3
```

## Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key

The tool will automatically prompt for your API key on first use and store it securely for future sessions.

## Output Formats

- `text` (default): Plain text transcription
- `srt`: SubRip subtitle format
- `vtt`: WebVTT subtitle format  
- `verbose_json`: Detailed JSON with timestamps and metadata

## License

MIT License
