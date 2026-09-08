package main

import "testing"

// Regression test for Issue 1: previously, main() set aofFileConfig.AppendDirName
// twice (once to --dir, immediately overwritten by --appenddirname) and never
// assigned aofFileConfig.Dir at all, so --dir silently had no effect on where the
// AOF directory was created. buildPersistenceConfig() is the extracted, testable
// form of that wiring.

func TestBuildPersistenceConfig_RDBUsesConfiguredDir(t *testing.T) {
	rdbCfg, _ := buildPersistenceConfig("/some/configured/path", "rdbfile.db", "yes", "appendonly", "appendonly.aof", "everysec")

	if rdbCfg.Dir != "/some/configured/path" {
		t.Fatalf("expected RDB dir %q, got %q", "/some/configured/path", rdbCfg.Dir)
	}
	if rdbCfg.DbFileName != "rdbfile.db" {
		t.Fatalf("expected RDB filename %q, got %q", "rdbfile.db", rdbCfg.DbFileName)
	}
}

func TestBuildPersistenceConfig_AOFUsesConfiguredDir(t *testing.T) {
	_, aofCfg := buildPersistenceConfig("/some/configured/path", "rdbfile.db", "yes", "appendonly", "appendonly.aof", "everysec")

	if aofCfg.Dir != "/some/configured/path" {
		t.Fatalf("expected AOF base dir %q, got %q (this is the Issue 1 regression: --dir must control the AOF directory too)", "/some/configured/path", aofCfg.Dir)
	}
	if aofCfg.AppendDirName != "appendonly" {
		t.Fatalf("expected append dir name %q, got %q", "appendonly", aofCfg.AppendDirName)
	}
	if aofCfg.AppendFilename != "appendonly.aof" {
		t.Fatalf("expected append filename %q, got %q", "appendonly.aof", aofCfg.AppendFilename)
	}
	if aofCfg.AppendOnly != "yes" {
		t.Fatalf("expected appendonly %q, got %q", "yes", aofCfg.AppendOnly)
	}
	if aofCfg.AppendFsync != "everysec" {
		t.Fatalf("expected appendfsync %q, got %q", "everysec", aofCfg.AppendFsync)
	}
}

func TestBuildPersistenceConfig_RDBAndAOFShareTheSameBaseDir(t *testing.T) {
	rdbCfg, aofCfg := buildPersistenceConfig("/data", "rdbfile.db", "yes", "appendonly", "appendonly.aof", "everysec")

	if rdbCfg.Dir != aofCfg.Dir {
		t.Fatalf("expected RDB dir (%q) and AOF base dir (%q) to match so persistence state lives under one configured directory", rdbCfg.Dir, aofCfg.Dir)
	}
}

func TestBuildPersistenceConfig_DifferentDirsDoNotShareState(t *testing.T) {
	_, aofCfgA := buildPersistenceConfig("/data/instance-a", "rdbfile.db", "yes", "appendonly", "appendonly.aof", "everysec")
	_, aofCfgB := buildPersistenceConfig("/data/instance-b", "rdbfile.db", "yes", "appendonly", "appendonly.aof", "everysec")

	if aofCfgA.Dir == aofCfgB.Dir {
		t.Fatalf("expected two different --dir values to produce two different AOF base dirs, both got %q", aofCfgA.Dir)
	}
}
