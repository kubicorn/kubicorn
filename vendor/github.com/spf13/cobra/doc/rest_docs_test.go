package doc

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
<<<<<<< HEAD
=======
	"strings"
>>>>>>> Initial dep workover
	"testing"

	"github.com/spf13/cobra"
)

func TestGenRSTDoc(t *testing.T) {
<<<<<<< HEAD
	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := GenReST(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	checkStringContains(t, output, echoCmd.Long)
	checkStringContains(t, output, echoCmd.Example)
	checkStringContains(t, output, "boolone")
	checkStringContains(t, output, "rootflag")
	checkStringContains(t, output, rootCmd.Short)
	checkStringContains(t, output, echoSubCmd.Short)
	checkStringOmits(t, output, deprecatedCmd.Short)
}

func TestGenRSTNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := GenReST(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	unexpected := "Auto generated"
	checkStringOmits(t, output, unexpected)
}

func TestGenRSTTree(t *testing.T) {
	c := &cobra.Command{Use: "do [OPTIONS] arg1 arg2"}

=======
	c := initializeWithRootCmd()
	// Need two commands to run the command alphabetical sort
	cmdEcho.AddCommand(cmdTimes, cmdEchoSub, cmdDeprecated)
	c.AddCommand(cmdPrint, cmdEcho)
	cmdRootWithRun.PersistentFlags().StringVarP(&flags2a, "rootflag", "r", "two", strtwoParentHelp)

	out := new(bytes.Buffer)

	// We generate on s subcommand so we have both subcommands and parents
	if err := GenReST(cmdEcho, out); err != nil {
		t.Fatal(err)
	}
	found := out.String()

	// Our description
	expected := cmdEcho.Long
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	// Better have our example
	expected = cmdEcho.Example
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	// A local flag
	expected = "boolone"
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	// persistent flag on parent
	expected = "rootflag"
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	// We better output info about our parent
	expected = cmdRootWithRun.Short
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	// And about subcommands
	expected = cmdEchoSub.Short
	if !strings.Contains(found, expected) {
		t.Errorf("Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	}

	unexpected := cmdDeprecated.Short
	if strings.Contains(found, unexpected) {
		t.Errorf("Unexpected response.\nFound: %v\nBut should not have!!\n", unexpected)
	}
}

func TestGenRSTNoTag(t *testing.T) {
	c := initializeWithRootCmd()
	// Need two commands to run the command alphabetical sort
	cmdEcho.AddCommand(cmdTimes, cmdEchoSub, cmdDeprecated)
	c.AddCommand(cmdPrint, cmdEcho)
	c.DisableAutoGenTag = true
	cmdRootWithRun.PersistentFlags().StringVarP(&flags2a, "rootflag", "r", "two", strtwoParentHelp)
	out := new(bytes.Buffer)

	if err := GenReST(c, out); err != nil {
		t.Fatal(err)
	}
	found := out.String()

	unexpected := "Auto generated"
	checkStringOmits(t, found, unexpected)

}

func TestGenRSTTree(t *testing.T) {
	cmd := &cobra.Command{
		Use: "do [OPTIONS] arg1 arg2",
	}
>>>>>>> Initial dep workover
	tmpdir, err := ioutil.TempDir("", "test-gen-rst-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %s", err.Error())
	}
	defer os.RemoveAll(tmpdir)

<<<<<<< HEAD
	if err := GenReSTTree(c, tmpdir); err != nil {
=======
	if err := GenReSTTree(cmd, tmpdir); err != nil {
>>>>>>> Initial dep workover
		t.Fatalf("GenReSTTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.rst")); err != nil {
		t.Fatalf("Expected file 'do.rst' to exist")
	}
}

func BenchmarkGenReSTToFile(b *testing.B) {
<<<<<<< HEAD
=======
	c := initializeWithRootCmd()
>>>>>>> Initial dep workover
	file, err := ioutil.TempFile("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
<<<<<<< HEAD
		if err := GenReST(rootCmd, file); err != nil {
=======
		if err := GenReST(c, file); err != nil {
>>>>>>> Initial dep workover
			b.Fatal(err)
		}
	}
}
