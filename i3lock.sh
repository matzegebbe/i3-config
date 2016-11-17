#!/bin/bash
scrot  'lockbg.png' -e 'convert -blur 0x4 $f ~/lockbg.png'
convert -gravity center -composite ~/lockbg.png ~/lock.png ~/lockfinal.png
i3lock -i ~/lockfinal.png
#i3lock -i ~/lockbg.png
rm ~/lockfinal.png ~/lockbg.png
