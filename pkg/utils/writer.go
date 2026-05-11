package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Writer struct {
	cfg *Config

	file *os.File
	buf  *bufio.Writer
	enc  *json.Encoder

	mu            sync.Mutex
	staticData    map[string]any
	dynamicData   map[string]any
	staticFlushed bool
}

func NewWriter(cfg *Config) *Writer {
	w := &Writer{
		cfg:         cfg,
		staticData:  make(map[string]any),
		dynamicData: make(map[string]any),
	}

	if cfg.OutputDir != "" {
		Debugf("writer: creating output dir=%q uuid=%q", cfg.OutputDir, cfg.UUID)
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			log.Printf("writer: failed to create output dir %s: %v", cfg.OutputDir, err)
		}

		path := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s.jsonl", cfg.UUID))
		f, err := os.Create(path)
		if err != nil {
			log.Printf("writer: failed to create %s: %v. Falling back to stdout.", path, err)
			w.buf = bufio.NewWriter(os.Stdout)
		} else {
			w.file = f
			w.buf = bufio.NewWriter(f)
			log.Printf("writer: output %s", path)
		}
	} else {
		Debugf("writer: no output dir, writing to stdout")
		w.buf = bufio.NewWriter(os.Stdout)
	}

	w.enc = json.NewEncoder(w.buf)
	return w
}

func (w *Writer) Static(name string, data any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.staticData[name] = data
	return nil
}

func (w *Writer) Dynamic(name string, data any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dynamicData[name] = data
	return nil
}

func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.staticFlushed && len(w.staticData) > 0 {
		Debugf("writer: flushing static data with %d sections", len(w.staticData))
		w.writeStatic()
		w.staticFlushed = true
	}

	if len(w.dynamicData) == 0 {
		return nil
	}

	snapshot := make(map[string]any)
	snapshot["timestamp"] = GetTimestamp()

	for name, data := range w.dynamicData {
		snapshot[name] = data
	}

	sections := make([]string, 0, len(w.dynamicData))
	for name := range w.dynamicData {
		sections = append(sections, name)
	}
	Debugf("writer: flushing dynamic data sections=%v", sections)

	w.dynamicData = make(map[string]any)

	var toWrite any = snapshot
	if w.cfg.Flatten {
		Debugf("writer: flattening dynamic data")
		toWrite = Flatten(snapshot)
	}

	if err := w.enc.Encode(toWrite); err != nil {
		return err
	}
	return w.buf.Flush()
}

func (w *Writer) writeStatic() {
	static := make(map[string]any)
	static["uuid"] = w.cfg.UUID
	static["timestamp"] = GetTimestamp()
	for name, data := range w.staticData {
		static[name] = data
	}

	sections := make([]string, 0, len(w.staticData))
	for name := range w.staticData {
		sections = append(sections, name)
	}
	Debugf("writer: writing static line uuid=%s sections=%v", w.cfg.UUID, sections)

	var toWrite any = static
	if w.cfg.Flatten {
		Debugf("writer: flattening static data")
		toWrite = Flatten(static)
	}

	if err := w.enc.Encode(toWrite); err != nil {
		log.Printf("writer: failed to write static line: %v", err)
		return
	}
	if err := w.buf.Flush(); err != nil {
		log.Printf("writer: failed to flush static line: %v", err)
	}
	log.Printf("writer: static metrics written")
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	Debugf("writer: closing (file=%v)", w.file != nil)
	if err := w.buf.Flush(); err != nil {
		Debugf("writer: final flush error: %v", err)
		return err
	}
	if w.file != nil {
		Debugf("writer: closing file %s", w.file.Name())
		return w.file.Close()
	}
	return nil
}
