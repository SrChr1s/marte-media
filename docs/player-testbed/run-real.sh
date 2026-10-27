#!/usr/bin/env bash
S=/tmp/claude-1000/-home-srchr1s-Satrn-marte-desc/8547ec3e-55bc-4b17-9b05-5f1a4d189503/scratchpad
P=$S/preal; LOG=$P/run.log; : > $LOG; rm -f $P/shots/*.png
before=$(niri msg windows | awk '/Window ID/{gsub(":","",$3); print $3}' | sort -n | tr "\n" " ")
orig=$(niri msg focused-window | awk '/Window ID/{gsub(":","",$3); print $3}')
XDG_CONFIG_HOME=$P/cfg PATH=/usr/bin:/bin MARTE_JS1="$(cat $P/scenario.js)" $S/realapp/marte-desc >$LOG 2>&1 &
app=$!; sleep 3
new=""; for w in $(niri msg windows | awk '/Window ID/{gsub(":","",$3); print $3}'); do case " $before " in *" $w "*) ;; *) new="$new $w";; esac; done
set -- $new; win=$1; [ -n "$win" ] && niri msg action focus-window --id "$win"
done_shots=""
for i in $(seq 1 700); do
  for name in $(grep -o "SHOT [0-9a-z_]*" $LOG | awk '{print $2}'); do
    case " $done_shots " in *" $name "*) ;; *)
      done_shots="$done_shots $name"; niri msg action screenshot-window --id "$win" --write-to-disk true >/dev/null 2>&1; sleep 0.7
      f=$(ls -t ~/Pictures/Screenshots | head -1); mv ~/Pictures/Screenshots/"$f" $P/shots/$name.png ;; esac
  done
  grep -q "\[real\] done" $LOG && break
  sleep 0.5
done
niri msg action focus-window --id "$orig" 2>/dev/null; kill $app 2>/dev/null; wait $app 2>/dev/null
grep -o "\[real\] .*" $LOG | grep -v SHOT
