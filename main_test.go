package main

import (
	"golang.org/x/tools/cover"
	"testing"
)

func TestMain_NewBlock(t *testing.T) {
	block := newBlock(1, 2, 3, 4)
	valid := block.StartLine == 1 &&
		block.StartCol == 2 &&
		block.EndLine == 3 &&
		block.EndCol == 4 &&
		block.NumStmt == 0 &&
		block.Count == 0

	if !valid {
		t.Error("block constructed correctly")
	}
}

func TestMain_IsWithin_SameBlock(t *testing.T) {
	if !isWithin(newBlock(1, 5, 10, 10), newBlock(1, 5, 10, 10)) {
		t.Error("block within itself")
	}
}

func TestMain_IsWithin_Lines(t *testing.T) {
	if !isWithin(newBlock(2, 5, 9, 5), newBlock(1, 5, 10, 5)) {
		t.Error("block within lines")
	}
}

func TestMain_IsWithin_OverlapStartCol(t *testing.T) {
	if isWithin(newBlock(1, 4, 10, 5), newBlock(1, 5, 10, 5)) {
		t.Error("block overlapping start column")
	}
}

func TestMain_IsWithin_OverlapEndCol(t *testing.T) {
	if isWithin(newBlock(1, 5, 10, 6), newBlock(1, 5, 10, 5)) {
		t.Error("block overlapping end column")
	}
}

func TestMain_IsWithin_OverlapStartLine(t *testing.T) {
	if isWithin(newBlock(1, 5, 10, 5), newBlock(2, 5, 10, 5)) {
		t.Error("block overlapping start line")
	}
}

func TestMain_IsWithin_OverlapEndLine(t *testing.T) {
	if isWithin(newBlock(1, 5, 11, 6), newBlock(1, 5, 10, 5)) {
		t.Error("block overlapping end line")
	}
}

func TestMain_IsWithin_DisjointLines(t *testing.T) {
	if isWithin(newBlock(1, 5, 5, 5), newBlock(6, 5, 10, 5)) {
		t.Error("disjoint lines")
	}
}

func TestMain_IsWithin_DisjointColumns(t *testing.T) {
	if isWithin(newBlock(1, 5, 5, 5), newBlock(5, 6, 10, 5)) {
		t.Error("disjoint lines")
	}
}

func TestMain_IsWithin_PrintProfile(t *testing.T) {
	profile := cover.Profile{
		FileName: "FILENAME",
		Mode:     "MODE",
		Blocks: []cover.ProfileBlock{
			*newBlock(1, 5, 10, 15),
			*newBlock(11, 5, 20, 15),
		},
	}

	out := printProfile(&profile)
	if out != "mode: MODE\nFILENAME:1.5,10.15 0 0\nFILENAME:11.5,20.15 0 0\n" {
		t.Error("print cover profile")
	}
}
