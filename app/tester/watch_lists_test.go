package tester

import "testing"

func list_watch_test(t *testing.T) {

	stage300_LPushInvalidatesWatch(t)
	stage301_RPushInvalidatesWatch(t)
	stage302_LPopInvalidatesWatch(t)
	// stage303_RPopInvalidatesWatch(t)

	stage304_LPushOtherKey(t)
	stage305_RPushOtherKey(t)
	// stage306_LPopOtherKey(t)
	// stage307_RPopOtherKey(t)
	stage308_LPushCreatesWatchedKey(t)
	stage309_RPushCreatesWatchedKey(t)
	stage310_LPushExistingWatchedKey(t) 
	stage311_RPushExistingWatchedKey(t) 
	stage312_LPushDifferentKey(t)
	stage313_RPushDifferentKey(t)
	stage314_LPopExistingWatchedKey(t)
	stage315_LPopMissingKey(t) 
	stage316_LPopEmptyListNoOp(t)

}



func stage300_LPushInvalidatesWatch(t *testing.T) {
	stage("STAGE 300: LPUSH INVALIDATES WATCH")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$5\r\nlistw\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$5\r\nlistw\r\n"+
			"$1\r\na\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$5\r\nlistw\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH invalidates WATCH")
}


func stage301_RPushInvalidatesWatch(t *testing.T) {
	stage("STAGE 301: RPUSH INVALIDATES WATCH")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$6\r\nlistww\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$6\r\nlistww\r\n"+
			"$1\r\na\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$6\r\nlistww\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH invalidates WATCH")
}


func stage302_LPopInvalidatesWatch(t *testing.T) {
	stage("STAGE 302: LPOP INVALIDATES WATCH")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$7\r\nlistwww\r\n"+
			"$1\r\nx\r\n")

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$7\r\nlistwww\r\n")

	send(conn2,
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$7\r\nlistwww\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$7\r\nlistwww\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPOP invalidates WATCH")
}



// func stage303_RPopInvalidatesWatch(t *testing.T) {
// 	stage("STAGE 303: RPOP INVALIDATES WATCH")

// 	conn1 := dial(t)
// 	conn2 := dial(t)

// 	defer conn1.Close()
// 	defer conn2.Close()

// 	send(conn1,
// 		"*3\r\n"+
// 			"$5\r\nRPUSH\r\n"+
// 			"$4\r\nlist\r\n"+
// 			"$1\r\nx\r\n")

// 	send(conn1,
// 		"*2\r\n"+
// 			"$5\r\nWATCH\r\n"+
// 			"$4\r\nlist\r\n")

// 	send(conn2,
// 		"*2\r\n"+
// 			"$4\r\nRPOP\r\n"+
// 			"$4\r\nlist\r\n")

// 	send(conn1,
// 		"*1\r\n"+
// 			"$5\r\nMULTI\r\n")

// 	send(conn1,
// 		"*2\r\n"+
// 			"$3\r\nGET\r\n"+
// 			"$4\r\nlist\r\n")

// 	resp := send(conn1,
// 		"*1\r\n"+
// 			"$4\r\nEXEC\r\n")

// 	expected := "*-1\r\n"

// 	if resp != expected {
// 		failf(t, "expected %q got %q", expected, resp)
// 	}

// 	pass("RPOP invalidates WATCH")
// }



func stage304_LPushOtherKey(t *testing.T) {
	stage("STAGE 304: LPUSH OTHER KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$7\r\nlist112\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$7\r\nlist222\r\n"+
			"$1\r\na\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH on other key does not invalidate WATCH")
}



func stage305_RPushOtherKey(t *testing.T) {
	stage("STAGE 305: RPUSH OTHER KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$7\r\nlist121\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$7\r\nlist221\r\n"+
			"$1\r\na\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH on other key does not invalidate WATCH")
}




// func stage307_RPopOtherKey(t *testing.T) {
// 	stage("STAGE 307: RPOP OTHER KEY")

// 	conn1 := dial(t)
// 	conn2 := dial(t)

// 	defer conn1.Close()
// 	defer conn2.Close()

// 	send(conn2,
// 		"*3\r\n"+
// 			"$5\r\nRPUSH\r\n"+
// 			"$5\r\nlist2\r\n"+
// 			"$1\r\nx\r\n")

// 	send(conn1,
// 		"*2\r\n"+
// 			"$5\r\nWATCH\r\n"+
// 			"$5\r\nlist1\r\n")

// 	send(conn2,
// 		"*2\r\n"+
// 			"$4\r\nRPOP\r\n"+
// 			"$5\r\nlist2\r\n")

// 	send(conn1,
// 		"*1\r\n"+
// 			"$5\r\nMULTI\r\n")

// 	send(conn1,
// 		"*1\r\n"+
// 			"$4\r\nPING\r\n")

// 	resp := send(conn1,
// 		"*1\r\n"+
// 			"$4\r\nEXEC\r\n")

// 	expected := "*1\r\n+PONG\r\n"

// 	if resp != expected {
// 		failf(t, "expected %q got %q", expected, resp)
// 	}

// 	pass("RPOP on other key does not invalidate WATCH")
// }




func stage308_LPushCreatesWatchedKey(t *testing.T) {
	stage("STAGE 300: LPUSH CREATES WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$27\r\nwatch_stage308_missing_list\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$27\r\nwatch_stage308_missing_list\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$27\r\nwatch_stage308_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH creating watched key aborts transaction")
}



func stage309_RPushCreatesWatchedKey(t *testing.T) {
	stage("STAGE 301: RPUSH CREATES WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$27\r\nwatch_stage309_missing_list\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$27\r\nwatch_stage309_missing_list\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$26\r\nwatch_stage309_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH creating watched key aborts transaction")
}


func stage310_LPushExistingWatchedKey(t *testing.T) {
	stage("STAGE 302: LPUSH EXISTING WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$28\r\nwatch_stage310_existing_list\r\n"+
			"$5\r\nfirst\r\n")

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$28\r\nwatch_stage310_existing_list\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$28\r\nwatch_stage310_existing_list\r\n"+
			"$6\r\nsecond\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$26\r\nwatch_stage310_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH on existing watched key aborts transaction")
}



func stage311_RPushExistingWatchedKey(t *testing.T) {
	stage("STAGE 303: RPUSH EXISTING WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$28\r\nwatch_stage311_existing_list\r\n"+
			"$5\r\nfirst\r\n")

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$28\r\nwatch_stage311_existing_list\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$28\r\nwatch_stage311_existing_list\r\n"+
			"$6\r\nsecond\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$26\r\nwatch_stage311_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH on existing watched key aborts transaction")
}



func stage312_LPushDifferentKey(t *testing.T) {
	stage("STAGE 304: LPUSH DIFFERENT KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$23\r\nwatch_stage312_list_one\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$23\r\nwatch_stage312_list_two\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$23\r\nwatch_stage312_list_one\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n:0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH on different key does not invalidate WATCH")
}



func stage313_RPushDifferentKey(t *testing.T) {
	stage("STAGE 305: RPUSH DIFFERENT KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$23\r\nwatch_stage313_list_one\r\n")

	send(conn2,
		"*3\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$23\r\nwatch_stage313_list_two\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$23\r\nwatch_stage313_list_one\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n:0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH on different key does not invalidate WATCH")
}



func stage314_LPopExistingWatchedKey(t *testing.T) {
	stage("STAGE 306: LPOP EXISTING WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*3\r\n"+
			"$5\r\nLPUSH\r\n"+
			"$28\r\nwatch_stage314_existing_list\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$28\r\nwatch_stage314_existing_list\r\n")

	send(conn2,
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$28\r\nwatch_stage314_existing_list\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$26\r\nwatch_stage314_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPOP invalidates WATCH")
}


func stage315_LPopMissingKey(t *testing.T) {
	stage("STAGE 307: LPOP MISSING KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$27\r\nwatch_stage314_missing_list\r\n")

	send(conn2,
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$27\r\nwatch_stage314_missing_list\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$4\r\nLLEN\r\n"+
			"$26\r\nwatch_stage314_other_list\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n:0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPOP on missing key does not invalidate WATCH")
}



func stage316_LPopEmptyListNoOp(t *testing.T) {
   stage("STAGE 308: LPOP EMPTY LIST NO-OP")

   conn1 := dial(t)
   conn2 := dial(t)
   defer conn1.Close()
   defer conn2.Close()

   // 1. Create a list with 1 item and pop it so the key is deleted/empty
   send(conn1, "*3\r\n$5\r\nLPUSH\r\n$4\r\nempt\r\n$1\r\nx\r\n")
   send(conn1, "*2\r\n$4\r\nLPOP\r\n$4\r\nempt\r\n")

   // 2. Watch the now-empty key
   send(conn1, "*2\r\n$5\r\nWATCH\r\n$4\r\nempt\r\n")

   // 3. Another client pops from the empty list (returns nil, does nothing)
   send(conn2, "*2\r\n$4\r\nLPOP\r\n$4\r\nempt\r\n")

   // 4. Transaction should still succeed!
   send(conn1, "*1\r\n$5\r\nMULTI\r\n")
   send(conn1, "*1\r\n$4\r\nPING\r\n")
   resp := send(conn1, "*1\r\n$4\r\nEXEC\r\n")

   expected := "*1\r\n+PONG\r\n"
   if resp != expected {
      failf(t, "expected %q got %q", expected, resp)
   }
   pass("LPOP on an emptied list does not invalidate WATCH")
}