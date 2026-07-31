package tester

import (
	"strings"
	"sync"
	"testing"
)

func auth_test(t *testing.T) {
	stageXX_AuthWrongArguments(t)
	stageXX_AuthNoPasswordSet(t)
	stageXX_AuthWrongPassword(t)
	stageXX_AuthCorrectPassword(t)
	stageXX_FailedAuthDoesNotAuthenticate(t)
	stageXX_AuthUsernamePasswordForm(t)
	stageXX_FailedReauthDoesNotDeauthenticate(t)
	stageXX_CommandsBlockedBeforeAuth(t)
	stageXX_ClearingPasswordReopensAccess(t)
	stageXX_ReauthWithCorrectPasswordIsIdempotent(t)
	stageXX_AuthConcurrent(t)
	stageXX_RemoveExistingPassword(t)
	stageXX_RemoveNonExistentPassword(t)
   stageXX_RemoveOneOfMultiplePasswords(t)

}

func stageXX_AuthWrongArguments(t *testing.T) {
	stage("AUTH: AUTH WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nAUTH\r\n")

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "AUTH with no arguments: expected RESP error, got %q", resp)
	}

	if !strings.Contains(strings.ToLower(msg), "wrong number of arguments") {
		failf(t,
			"AUTH with no arguments: expected error containing %q, got %q",
			"wrong number of arguments",
			msg,
		)
	}

	resp = send(conn,
		"*4\r\n"+
			"$4\r\nAUTH\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n")

	msg, ok = parseError(resp)
	if !ok {
		failf(t, "AUTH with too many arguments: expected RESP error, got %q", resp)
	}

	if !strings.Contains(strings.ToLower(msg), "wrong number of arguments") {
		failf(t,
			"AUTH with too many arguments: expected error containing %q, got %q",
			"wrong number of arguments",
			msg,
		)
	}

	pass("wrong argument counts rejected")
}

func stageXX_AuthNoPasswordSet(t *testing.T) {
	stage("AUTH: AUTH WHEN NO PASSWORD IS CONFIGURED")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nwhatever\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "AUTH with no password configured: expected RESP error, got %q", resp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t,
			"AUTH with no password configured: expected WRONGPASS error, got %q",
			msg,
		)
	}

	pass("AUTH correctly rejected when no password is configured")
}
func stageXX_AuthWrongPassword(t *testing.T) {
	stage("AUTH: AUTH WRONG PASSWORD")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret89\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	resp = send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$5\r\nwrong\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "AUTH with wrong password: expected RESP error, got %q", resp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t,
			"AUTH with wrong password: expected WRONGPASS, got %q",
			msg,
		)
	}

	resp= send(conn,
			"*4\r\n"+
				"$3\r\nACL\r\n"+
				"$7\r\nSETUSER\r\n"+
				"$7\r\ndefault\r\n"+
				"$6\r\nnopass\r\n",
		)

		if resp != "+OK\r\n" {
			failf(t, "cleanup: expected +OK from ACL SETUSER nopass, got %q", resp)
		}

	pass("wrong password rejected")
}

func stageXX_AuthCorrectPassword(t *testing.T) {
	stage("AUTH: AUTH CORRECT PASSWORD")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
			"*4\r\n"+
				"$3\r\nACL\r\n"+
				"$7\r\nSETUSER\r\n"+
				"$7\r\ndefault\r\n"+
				"$9\r\n>secret89\r\n",
		)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	resp = send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret89\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK, got %q", resp)
	}

	resp = send(conn,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	if resp != "+PONG\r\n" {
		failf(t, "expected +PONG after AUTH, got %q", resp)
	}


	resp = send(conn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK from ACL SETUSER nopass, got %q", resp)
	}

	pass("correct password authenticated client")
}


func stageXX_FailedAuthDoesNotAuthenticate(t *testing.T) {
	stage("AUTH: FAILED AUTH DOES NOT AUTHENTICATE")

	// Configure the password on an already-authenticated connection.
	setupConn := dial(t)
	defer setupConn.Close()

	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret89\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	// Open a new connection. This one should NOT be authenticated.
	conn := dial(t)
	defer conn.Close()

	resp = send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$5\r\nwrong\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "expected WRONGPASS, got %q", resp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t, "expected WRONGPASS, got %q", msg)
	}

	resp = send(conn,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	msg, ok = parseError(resp)
	if !ok {
		failf(t, "expected NOAUTH, got %q", resp)
	}

	if !strings.Contains(msg, "NOAUTH") {
		failf(t, "expected NOAUTH after failed AUTH, got %q", msg)
	}

  //reset everything

	resp = send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret89\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from AUTH, got %q", resp)
	}


	resp = send(conn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK from ACL SETUSER nopass, got %q", resp)
	}

	pass("failed AUTH did not authenticate client")
}

func stageXX_AuthUsernamePasswordForm(t *testing.T) {
	stage("AUTH: AUTH WITH USERNAME + PASSWORD FORM")

	// Configure the default user with a password.
	setupConn := dial(t)
	defer setupConn.Close()


	resp:= send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret90\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	// Brand-new connection should require authentication.
	conn := dial(t)
	defer conn.Close()

	authResp := send(conn,
		"*3\r\n"+
			"$4\r\nAUTH\r\n"+
			"$7\r\ndefault\r\n"+
			"$8\r\nsecret90\r\n",
	)

	if authResp != "+OK\r\n" {
		failf(t, "expected +OK for AUTH <default> <password>, got %q", authResp)
	}

	pingResp := send(conn,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	if pingResp != "+PONG\r\n" {
		failf(t, "expected +PONG after successful AUTH, got %q", pingResp)
	}

	// Cleanup: remove the password.
	resp = send(conn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)


	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK removing password, got %q", resp)
	}

	pass("AUTH <username> <password> form accepted")
}

func stageXX_FailedReauthDoesNotDeauthenticate(t *testing.T) {
	stage("AUTH: FAILED RE-AUTH DOES NOT DEAUTHENTICATE AN ALREADY-AUTHENTICATED CLIENT")

	// Reset to a known state.
	setupConn := dial(t)
	defer setupConn.Close()


	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret91\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	conn := dial(t)
	defer conn.Close()

	authResp := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret91\r\n",
	)

	if authResp != "+OK\r\n" {
		failf(t, "expected +OK for correct password, got %q", authResp)
	}

	badAuthResp := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$5\r\nwrong\r\n",
	)

	msg, ok := parseError(badAuthResp)
	if !ok {
		failf(t, "expected WRONGPASS, got %q", badAuthResp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t, "expected WRONGPASS, got %q", msg)
	}

	// The previous successful AUTH should still be in effect.
	pingResp := send(conn,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	if pingResp != "+PONG\r\n" {
		failf(t, "expected still-authenticated session to work after failed re-auth, got %q", pingResp)
	}

	// Cleanup.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK from ACL SETUSER, got %q", resp)
	}

	pass("session remained authenticated despite a failed re-auth attempt")
}
func stageXX_CommandsBlockedBeforeAuth(t *testing.T) {
	stage("AUTH: NON-PING COMMANDS ALSO BLOCKED BEFORE AUTH")

	setupConn := dial(t)
	defer setupConn.Close()


	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret92\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	conn := dial(t)
	defer conn.Close()

	resp = send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$8\r\nauth-key\r\n"+
			"$3\r\nval\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "expected NOAUTH, got %q", resp)
	}

	if !strings.Contains(msg, "NOAUTH") {
		failf(t, "expected NOAUTH before AUTH, got %q", msg)
	}
  
	authResp := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret92\r\n",
	)

	if authResp != "+OK\r\n" {
		failf(t, "expected +OK for correct password, got %q", authResp)
	}

	setOK := send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$8\r\nauth-key\r\n"+
			"$3\r\nval\r\n",
	)

	if setOK != "+OK\r\n" {
		failf(t, "expected +OK for SET after AUTH, got %q", setOK)
	}

	// Cleanup.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK from ACL SETUSER, got %q", resp)
	}

	pass("data commands blocked pre-auth, allowed post-auth")
}
func stageXX_ClearingPasswordReopensAccess(t *testing.T) {
	stage("AUTH: CLEARING PASSWORD REOPENS ACCESS FOR NEW CONNECTIONS")

	connA := dial(t)
	defer connA.Close()

	resp := send(connA,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret93\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	authResp := send(connA,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret93\r\n",
	)

	if authResp != "+OK\r\n" {
		failf(t, "expected +OK for AUTH, got %q", authResp)
	}

	// Re-enable nopass.
	resp = send(connA,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK restoring nopass, got %q", resp)
	}

	connB := dial(t)
	defer connB.Close()

	pingResp := send(connB,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	if pingResp != "+PONG\r\n" {
		failf(t, "expected new connection to work after restoring nopass, got %q", pingResp)
	}

	pass("restoring nopass reopened unauthenticated access")
}

func stageXX_ReauthWithCorrectPasswordIsIdempotent(t *testing.T) {
	stage("STAGE 94: RE-AUTH WITH CORRECT PASSWORD ON ALREADY-AUTHENTICATED CONNECTION")

	setupConn := dial(t)
	defer setupConn.Close()

	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret94\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	conn := dial(t)
	defer conn.Close()

	firstAuth := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret94\r\n",
	)

	if firstAuth != "+OK\r\n" {
		failf(t, "expected +OK for first AUTH, got %q", firstAuth)
	}

	secondAuth := send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret94\r\n",
	)

	if secondAuth != "+OK\r\n" {
		failf(t, "expected +OK for second AUTH, got %q", secondAuth)
	}

	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK restoring nopass, got %q", resp)
	}

	pass("re-authenticating with the correct password succeeded")
}
func stageXX_AuthConcurrent(t *testing.T) {
	stage("AUTH: CONCURRENT AUTH CLIENTS")

	setupConn := dial(t)
	defer setupConn.Close()


	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret95\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			authResp := send(conn,
				"*2\r\n"+
					"$4\r\nAUTH\r\n"+
					"$8\r\nsecret95\r\n",
			)

			if authResp != "+OK\r\n" {
				failf(t, "client %d: expected +OK from AUTH, got %q", i, authResp)
				return
			}

			pingResp := send(conn,
				"*1\r\n"+
					"$4\r\nPING\r\n",
			)

			if pingResp != "+PONG\r\n" {
				failf(t, "client %d: expected +PONG after AUTH, got %q", i, pingResp)
			}
		}(i)
	}

	wg.Wait()

	// Cleanup.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK restoring nopass, got %q", resp)
	}

	pass("50 concurrent clients authenticated successfully with no race or deadlock")
}

func stageXX_RemoveExistingPassword(t *testing.T) {
	stage("AUTH: REMOVE EXISTING PASSWORD")

	setupConn := dial(t)
	defer setupConn.Close()

	// Reset to a known state.
	resp := send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK from ACL SETUSER, got %q", resp)
	}

	// Add a password.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n>secret96\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK adding password, got %q", resp)
	}

	// Authenticate successfully.
	conn := dial(t)
	defer conn.Close()

	resp = send(conn,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret96\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK from AUTH, got %q", resp)
	}

	// Remove the password.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$9\r\n<secret96\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK removing password, got %q", resp)
	}

	// New connection should no longer authenticate with it.
	conn2 := dial(t)
	defer conn2.Close()

	resp = send(conn2,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$8\r\nsecret96\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "expected WRONGPASS, got %q", resp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t, "expected WRONGPASS, got %q", msg)
	}

	resp = send(conn2,
		"*1\r\n"+
			"$4\r\nPING\r\n",
	)

	msg, ok = parseError(resp)
	if !ok {
		failf(t, "expected NOAUTH, got %q", resp)
	}

	if !strings.Contains(msg, "NOAUTH") {
		failf(t, "expected NOAUTH, got %q", msg)
	}

	// Restore default state.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK restoring nopass, got %q", resp)
	}

	pass("existing password removed successfully")
}

func stageXX_RemoveNonExistentPassword(t *testing.T) {
	stage("AUTH: REMOVE NON-EXISTENT PASSWORD")

	conn := dial(t)
	defer conn.Close()

	// Removing a password that doesn't exist should be a no-op.
	resp := send(conn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$15\r\n<does-not-exist\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK removing a non-existent password, got %q", resp)
	}

	pass("removing a non-existent password succeeded")
}


func stageXX_RemoveOneOfMultiplePasswords(t *testing.T) {
	stage("AUTH: REMOVE ONE OF MULTIPLE PASSWORDS")

	setupConn := dial(t)
	defer setupConn.Close()


	// Add two passwords.
	resp := send(setupConn,
		"*5\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$4\r\n>foo\r\n"+
			"$4\r\n>bar\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "setup: expected +OK adding passwords, got %q", resp)
	}

	// Remove only "foo".
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$4\r\n<foo\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK removing password, got %q", resp)
	}

	// "foo" should no longer authenticate.
	conn1 := dial(t)
	defer conn1.Close()

	resp = send(conn1,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$3\r\nfoo\r\n",
	)

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "expected WRONGPASS for removed password, got %q", resp)
	}

	if !strings.Contains(msg, "WRONGPASS") {
		failf(t, "expected WRONGPASS, got %q", msg)
	}

	// "bar" should still authenticate.
	conn2 := dial(t)
	defer conn2.Close()

	resp = send(conn2,
		"*2\r\n"+
			"$4\r\nAUTH\r\n"+
			"$3\r\nbar\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK authenticating with remaining password, got %q", resp)
	}

	// Cleanup.
	resp = send(setupConn,
		"*4\r\n"+
			"$3\r\nACL\r\n"+
			"$7\r\nSETUSER\r\n"+
			"$7\r\ndefault\r\n"+
			"$6\r\nnopass\r\n",
	)

	if resp != "+OK\r\n" {
		failf(t, "cleanup: expected +OK restoring nopass, got %q", resp)
	}

	pass("only the requested password was removed")
}