#!/bin/tcsh
#
if ( `id -u` != 0 ) then
  echo must be run as root
  exit 1
endif
# 128.114.46.26 bobdev tap2 gdev

if ( $# < 4 ) then
  echo usage create.csh <jid> <int> <ipaddr> <type>
  exit 1
endif

set jid = $1
set t = $2
set i = $3
set type = $4

ifconfig bridge46 | grep $t
# network
if ( $? != 0 ) then
  echo no tap interface
  exit 1
endif

set re = 13.2-RELEASE
echo creating jail...

bastille create $jid $re $i/24 $t
# copy authkeys
cp /tmp/authkey /bas-stor/bastille/templates/bob/ssh-root/files/authorized_keys
rm /tmp/authkey
bastille template $jid bob/ssh-root
bastille template $jid bob/$type

echo Done


