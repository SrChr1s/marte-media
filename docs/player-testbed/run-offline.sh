#!/usr/bin/env bash
# run.sh : serve test media, run the app with the fake yt-dlp and the scenario, screenshot at each SHOT marker
# S: working folder holding media/, pdemo|preal/ (scenario.js, cfg, bin) and realapp/ (see README.md here)
S=${S:-/tmp/marte-testbed}
REPO=${REPO:-$(cd "$(dirname "$0")/../.." && pwd)}
P=$S/pdemo; LOG=$P/run.log; : > $LOG; rm -f $P/shots/*.png
python3 $S/media/rangeserver.py 18765 & srv=$!
before=$(niri msg windows | awk '/Window ID/{gsub(":","",$3); print $3}' | sort -n | tr "\n" " ")
orig=$(niri msg focused-window | awk '/Window ID/{gsub(":","",$3); print $3}')
XDG_CONFIG_HOME=$P/cfg PATH=$P/bin:$PATH MARTE_JS1="$(cat $P/scenario.js)" $REPO/build/bin/marte-media >$LOG 2>&1 &
app=$!
sleep 3
new=""; for w in $(niri msg windows | awk '/Window ID/{gsub(":","",$3); print $3}'); do case " $before " in *" $w "*) ;; *) new="$new $w";; esac; done
set -- $new; win=$1
[ -n "$win" ] && niri msg action focus-window --id "$win"
done_shots=""
for i in $(seq 1 600); do
  for name in $(grep -o "SHOT [0-9a-z_]*" $LOG | awk '{print $2}'); do
    case " $done_shots " in *" $name "*) ;; *)
      done_shots="$done_shots $name"
      niri msg action screenshot-window --id "$win" --write-to-disk true >/dev/null 2>&1; sleep 0.7
      f=$(ls -t ~/Pictures/Screenshots | head -1); mv ~/Pictures/Screenshots/"$f" $P/shots/$name.png ;;
    esac
  done
  grep -q "\[pl\] done" $LOG && break
  sleep 0.5
done
niri msg action focus-window --id "$orig" 2>/dev/null
kill $app $srv 2>/dev/null; wait $app $srv 2>/dev/null
grep -o "\[pl\] .*" $LOG | grep -v "SHOT"
