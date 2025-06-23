# Pindar - Audio Transcription CLI Tool

Pindar is a Go-based command-line tool that transcribes audio files using the OpenAI API.

## Features

- Transcribe audio files in various formats (flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm)
- Support for different output formats (text, srt, vtt, verbose_json)
- Option to save output to a file or print to stdout
- Custom language specification
- Temperature adjustment for transcription randomness

## Installation

1. Make sure you have Go installed (1.18+)
2. Clone the repository
3. Build the binary:

```bash
go build -o pindar
```

## Usage

```bash
# Basic usage
./pindar /path/to/audio/file.mp3

# With OpenAI API key (if not set in environment)
./pindar --api-key=your_api_key /path/to/audio/file.mp3

# Save output to a specific directory
./pindar --output-dir=transcriptions /path/to/audio/file.mp3

# Specify the output format
./pindar --format=srt /path/to/audio/file.mp3

# Specify the language to improve accuracy
./pindar --language=en /path/to/audio/file.mp3

# Use a different model (default is whisper-1)
./pindar --model=gpt-4o-transcribe /path/to/audio/file.mp3
```

## Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key

## API Key Configuration

Pindar uses the following priority order to find your OpenAI API key:

1. **Command-line argument**: `--api-key=your_key`
2. **Environment variable**: `OPENAI_API_KEY`
3. **Configuration file**: Stored in platform-specific config directory
4. **Interactive prompt**: If none found, you'll be prompted to enter it

When you enter an API key via the interactive prompt, it will be securely saved to:
- **macOS**: `~/Library/Application Support/pindar/config.json`
- **Linux**: `~/.config/pindar/config.json`
- **Windows**: `%APPDATA%\pindar\config.json`

The stored API key will be used for future runs unless overridden by a command-line argument or environment variable.

## Supported Audio Formats

- flac
- mp3
- mp4
- mpeg
- mpga
- m4a
- ogg
- wav
- webm

## Output Formats

- `text` (default): Plain text transcription
- `srt`: SubRip subtitle format
- `vtt`: WebVTT subtitle format
- `verbose_json`: Detailed JSON with additional information
- `json`: Simple JSON format (required for gpt-4o models)

## Limitations

- Maximum audio file size: 25MB
- For gpt-4o-transcribe and gpt-4o-mini-transcribe models, only 'json' response format is supported

## License

MIT
