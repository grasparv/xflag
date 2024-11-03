package xflag

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	XFlag     string         `xflag:"testcmd,Test command"`
	FlagInt   *int           `xflag:"10,An integer flag with a default of 10"`
	FlagInt2  *int           `xflag:"An integer flag without default"`
	FlagBool  *bool          `xflag:"false,A boolean flag"`
	FlagBool2 *bool          `xflag:"A boolean flag without default"`
	FlagDur   *time.Duration `xflag:"5s,A duration flag with a default of 5s"`
	FlagDur2  *time.Duration `xflag:"A duration flag without default"`
	Arg1      string         `xflag:"The first required argument"`
	Arg2      string         `xflag:"The second required argument"`
	Arg3      string         `xflag:"The third required argument with a default of 'fluff'"`
}

func TestParseOverview(t *testing.T) {
	_, err := Parse([]interface{}{TestStruct{}}, []string{"cmd", "testcmd"})
	require.Error(t, err)
}

func TestParseWithAllFlagsAndArguments(t *testing.T) {
	cmd, err := Parse([]interface{}{TestStruct{}}, []string{"cmd", "testcmd", "-flag-int", "42", "-flag-int2", "84", "-flag-bool", "true", "-flag-bool2", "false", "-flag-dur", "10s", "-flag-dur2", "30s", "arg1_value", "arg2_value", "arg3_value"})
	require.NoError(t, err, "Expected no error for command with all flags and arguments")

	parsedCmd := cmd.(*TestStruct)
	require.Equal(t, 42, *parsedCmd.FlagInt, "Expected FlagInt to be 42")
	require.Equal(t, 84, *parsedCmd.FlagInt2, "Expected FlagInt2 to be 84")
	require.Equal(t, true, *parsedCmd.FlagBool, "Expected FlagBool to be true")
	require.Equal(t, false, *parsedCmd.FlagBool2, "Expected FlagBool2 to be true")
	require.Equal(t, 10*time.Second, *parsedCmd.FlagDur, "Expected FlagDur to be 10s")
	require.Equal(t, 30*time.Second, *parsedCmd.FlagDur2, "Expected FlagDur2 to be 30s")
	require.Equal(t, "arg1_value", parsedCmd.Arg1, "Expected Arg1 to be 'arg1_value'")
	require.Equal(t, "arg2_value", parsedCmd.Arg2, "Expected Arg2 to be 'arg2_value'")
	require.Equal(t, "arg3_value", parsedCmd.Arg3, "Expected Arg3 to be 'arg3_value'")
}

func TestParseWithMissingMandatoryArgument(t *testing.T) {
	_, err := Parse([]interface{}{TestStruct{}}, []string{"cmd", "testcmd", "-flag-int", "42"})
	require.Error(t, err, "Expected an error for missing mandatory arguments")
	require.Contains(t, err.Error(), "usage:", "Expected error to mention missing required arguments")
}

func TestParseWithPartialFlagsAndArguments(t *testing.T) {
	cmd, err := Parse([]interface{}{TestStruct{}}, []string{"cmd", "testcmd", "-flag-int", "42", "requiredArg1", "requiredArg2", "fluff"})
	require.NoError(t, err, "Expected no error with partial flags and required arguments")

	parsedCmd := cmd.(*TestStruct)
	require.Equal(t, 42, *parsedCmd.FlagInt, "Expected FlagInt to be 42")
	require.Nil(t, parsedCmd.FlagInt2, "Expected FlagInt2 to be nil")
	require.Equal(t, false, *parsedCmd.FlagBool, "Expected FlagBool to default to false")
	require.Nil(t, parsedCmd.FlagBool2, "Expected FlagBool2 to be nil")
	require.Equal(t, 5*time.Second, *parsedCmd.FlagDur, "Expected FlagDur to default to 5s")
	require.Nil(t, parsedCmd.FlagDur2, "Expected FlagDur2 to be nil")
	require.Equal(t, "requiredArg1", parsedCmd.Arg1, "Expected Arg1 to be 'requiredArg1'")
	require.Equal(t, "requiredArg2", parsedCmd.Arg2, "Expected Arg2 to be 'requiredArg2'")
	require.Equal(t, "fluff", parsedCmd.Arg3, "Expected Arg3 to default to 'fluff'")
}

func TestParseInvalidCommand(t *testing.T) {
	_, err := Parse([]interface{}{TestStruct{}}, []string{"cmd", "invalidcmd"})
	require.Error(t, err, "Expected an error for an invalid command")
	require.Contains(t, err.Error(), "unrecognized command", "Expected error to mention unknown command")
}
