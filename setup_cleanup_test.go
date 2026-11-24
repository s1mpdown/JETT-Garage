package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Generating a setup for a track replaces the older Jett setup for the
// same slot (Q or R) — even from another car — while the other slot,
// non-Jett files and the file just written stay put. Replaced files are
// moved into JettBackups, not deleted.
func TestReplaceOlderSetups(t *testing.T) {
	root := t.TempDir()
	trackDir := filepath.Join(root, "UserData", "player", "Settings", "Imola")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(trackDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("Jett_AMG_IMOLA_Q.svm") // the file just generated
	write("Jett_RCF_IMOLA_Q.svm") // stale quali from another car: must move
	write("Jett_RCF_IMOLA_R.svm") // race slot: must stay
	write("MyOwnSetup.svm")       // not app-generated: must stay
	write("Jett_AMG_IMOLA_R.svm") // race slot: must stay

	dest := filepath.Join(trackDir, "Jett_AMG_IMOLA_Q.svm")
	moved := replaceOlderSetups(dest)
	if len(moved) != 1 || moved[0] != "Jett_RCF_IMOLA_Q.svm" {
		t.Fatalf("moved = %v, want [Jett_RCF_IMOLA_Q.svm]", moved)
	}
	if _, err := os.Stat(filepath.Join(trackDir, "Jett_RCF_IMOLA_Q.svm")); !os.IsNotExist(err) {
		t.Errorf("stale quali setup still in track folder")
	}
	bakDir := filepath.Join(root, "UserData", "player", "JettBackups", "Imola")
	ents, err := os.ReadDir(bakDir)
	if err != nil || len(ents) != 1 {
		t.Fatalf("JettBackups/Imola = %v, %v; want exactly the moved file", ents, err)
	}
	for _, stay := range []string{"Jett_AMG_IMOLA_Q.svm", "Jett_RCF_IMOLA_R.svm", "MyOwnSetup.svm", "Jett_AMG_IMOLA_R.svm"} {
		if _, err := os.Stat(filepath.Join(trackDir, stay)); err != nil {
			t.Errorf("%s should have stayed: %v", stay, err)
		}
	}
	// A race generation then replaces the older race setups only.
	write("Jett_F296_IMOLA_R.svm")
	moved = replaceOlderSetups(filepath.Join(trackDir, "Jett_F296_IMOLA_R.svm"))
	if len(moved) != 2 {
		t.Fatalf("race replace moved = %v, want the two older race setups", moved)
	}
}

// v1.3.4 naming: generated files are JETT_{CAR}-Q / JETT_{CAR}-R. A new
// generation must replace stale setups from BOTH naming generations
// (new JETT_…-Q and legacy Jett_…_Q) in the same slot, while the other
// slot and the user's own files stay put.
func TestReplaceOlderSetupsNewGeneration(t *testing.T) {
	root := t.TempDir()
	trackDir := filepath.Join(root, "UserData", "player", "Settings", "Barcelona")
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(trackDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("JETT_AMG-Q.svm")       // the file just generated
	write("JETT_RCF-Q.svm")       // stale new-generation quali: must move
	write("Jett_RCF_IMOLA_Q.svm") // stale legacy quali: must move
	write("JETT_RCF-R.svm")       // race slot: must stay
	write("MyOwnSetup.svm")       // not app-generated: must stay

	dest := filepath.Join(trackDir, "JETT_AMG-Q.svm")
	moved := replaceOlderSetups(dest)
	if len(moved) != 2 {
		t.Fatalf("moved = %v, want the two stale quali setups", moved)
	}
	got := map[string]bool{}
	for _, m := range moved {
		got[m] = true
	}
	for _, want := range []string{"JETT_RCF-Q.svm", "Jett_RCF_IMOLA_Q.svm"} {
		if !got[want] {
			t.Errorf("moved = %v, want %s among them", moved, want)
		}
		if _, err := os.Stat(filepath.Join(trackDir, want)); !os.IsNotExist(err) {
			t.Errorf("stale quali setup %s still in track folder", want)
		}
	}
	for _, stay := range []string{"JETT_AMG-Q.svm", "JETT_RCF-R.svm", "MyOwnSetup.svm"} {
		if _, err := os.Stat(filepath.Join(trackDir, stay)); err != nil {
			t.Errorf("%s should have stayed: %v", stay, err)
		}
	}
	bakDir := filepath.Join(root, "UserData", "player", "JettBackups", "Barcelona")
	ents, err := os.ReadDir(bakDir)
	if err != nil || len(ents) != 2 {
		t.Fatalf("JettBackups/Barcelona = %v, %v; want exactly the two moved files", ents, err)
	}
}
