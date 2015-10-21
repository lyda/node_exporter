#!/bin/bash

jails=0
for jail in $(fail2ban-client status | grep 'Jail list:' \
              | sed 's/.*Jail list:[\t ]*//;s/,//g'); do
  jails=$(( $jails + 1 ))
  fail2ban-client status $jail | awk -F: '
    $1 ~ /Currently failed/ {
      print "fail2ban_failed_current{jail=\"'"$jail"'\"} " $2;
    }
    $1 ~ /Total failed/ {
      print "fail2ban_failed_total{jail=\"'"$jail"'\"} " $2;
    }
    $1 ~ /Currently banned/ {
      print "fail2ban_banned_current{jail=\"'"$jail"'\"} " $2;
    }
    $1 ~ /Total banned/ {
      print "fail2ban_banned_total{jail=\"'"$jail"'\"} " $2;
    }'
done | sort | awk '
  BEGIN {failc=1; failt=1; banc=1; bant=1}
  /^fail2ban_failed_current/ { if (failc) {
      print "# HELP fail2ban_failed_current Current number of failures.";
      print "# TYPE fail2ban_failed_current gauge";
      failc=0;
    } }
  /^fail2ban_failed_total/ { if (failt) {
      print "# HELP fail2ban_failed_total Total number of failures.";
      print "# TYPE fail2ban_failed_total counter";
      failt=0;
    } }
  /^fail2ban_banned_current/ { if (banc) {
      print "# HELP fail2ban_banned_current Current number banned.";
      print "# TYPE fail2ban_banned_current gauge";
      banc=0;
    } }
  /^fail2ban_banned_total/ { if (bant) {
      print "# HELP fail2ban_banned_total Total number banned.";
      print "# TYPE fail2ban_banned_total counter";
      bant=0;
    } }
  { print $0 }'
cat << EOF
# HELP fail2ban_banned_total Total number banned.
# TYPE fail2ban_banned_total counter
fail2ban_jails $jails
EOF
