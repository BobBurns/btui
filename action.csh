#!/bin/tcsh

if ( $# < 2 ) then
  echo usage <action.csh> jid action
endif

set j = $1
set a = $2

echo action $a on $j
if ( $a == "list" ) then
  bastille list -a
  exit $?
endif

if ( $a == "destroy" ) then
  bastille destroy -f $j
  exit $?
endif

bastille $a $j
