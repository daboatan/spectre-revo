package main

import (
	"errors"
	"io"
	"testing"
)

type trackingWriteCloser struct {
	closed   bool
	closeErr error
}

func (w *trackingWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (w *trackingWriteCloser) Close() error {
	w.closed = true
	return w.closeErr
}

type trackingPasteStore struct {
	writer      *trackingWriteCloser
	saveCalled  bool
	savedClosed bool
	saveErr     error
}

func (s *trackingPasteStore) GenerateNewPasteID(bool) (PasteID, error) { return "test", nil }
func (s *trackingPasteStore) New(bool) (*Paste, error)                 { return nil, nil }
func (s *trackingPasteStore) Get(PasteID, []byte) (*Paste, error)      { return nil, nil }
func (s *trackingPasteStore) Save(*Paste) error {
	s.saveCalled = true
	s.savedClosed = s.writer.closed
	return s.saveErr
}
func (s *trackingPasteStore) Destroy(*Paste) error { return nil }
func (s *trackingPasteStore) EncryptionKeyForPasteWithPassword(*Paste, string) []byte {
	return nil
}
func (s *trackingPasteStore) readStream(*Paste) (*PasteReader, error) { return nil, nil }
func (s *trackingPasteStore) writeStream(p *Paste) (*PasteWriter, error) {
	return &PasteWriter{WriteCloser: s.writer, paste: p}, nil
}

func TestPasteWriterCloseFlushesBeforeSave(t *testing.T) {
	store := &trackingPasteStore{writer: &trackingWriteCloser{}}
	paste := &Paste{store: store}
	w, err := paste.Writer()
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if !store.saveCalled || !store.savedClosed {
		t.Fatal("paste was not saved after its body was closed")
	}
}

func TestPasteWriterClosePropagatesErrors(t *testing.T) {
	closeErr := errors.New("close failed")
	store := &trackingPasteStore{writer: &trackingWriteCloser{closeErr: closeErr}}
	paste := &Paste{store: store}
	w, _ := paste.Writer()
	if err := w.Close(); !errors.Is(err, closeErr) {
		t.Fatalf("Close() error = %v, want %v", err, closeErr)
	}
	if store.saveCalled {
		t.Fatal("paste metadata was saved after its body failed to close")
	}

	store.writer = &trackingWriteCloser{}
	store.saveErr = io.ErrUnexpectedEOF
	w, _ = paste.Writer()
	if err := w.Close(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Close() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}
