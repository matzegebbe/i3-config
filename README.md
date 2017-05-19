# i3-config

Here I share my i3-config with you!

For all nice features I use https://bitbucket.org/s_l_teichmann/i3stuff/

i3stuff
=======

Some helpers for my i3 window manager setup.

i3win.go
--------

Creates a dmenu choice list to select a window and focus it.
It scans the scrachpad and floating windows, too.

Similiar to quickswitch-for-i3 [1] but simpler
and much faster (compiles to a native code standalone binary).

Use with something like in your .i3/config

  bindsym exec $mod+g i3win

To influence the call of dmenu you can use the 
following options:

  -d=/path/to/dmenu
  -dmenu=/path/to/dmenu # path defaults to 'dmenu'

  -da=args
  -dmenu-args=args # args defaults to '-l 20 -b -i'

Build:

  $ go get bitbucket.org/s_l_teichmann/i3stuff/cmd/i3win

  Manual build:

  $ go get -u github.com/proxypoke/i3ipc
  $ hg clone https://s_l_teichmann@bitbucket.org/s_l_teichmann/i3stuff
  $ cd i3stuff/cmd/i3win
  $ go build

  Place the resulting i3win somewhere in your path.

[1] https://github.com/proxypoke/quickswitch-for-i3


i3switch.go
-----------

I enjoy having named workspaces in i3 like 'mail', 'web', 'work' 
and so on and I want easy switching between them.

The built-in solution is to use number prefixed workspace names or 
do an exact match search for a workspace. But why should I type
'mail' if there is only one workspace with a 'm' inside the name?
Or why should I confuse myself with re-numbering if the workspace
is moved to a different display over day?

i3switch.go does a partial and case insensitive search for a workspace and
switches to it if its found. If the binary is named i3move
the currently selected container is moved to the found workspace.
If the binary is named i3follow the currently container is moved
to the found workspace and the focus will follow.
If the workspace is not found a new workspace is created.

  $ go get bitbucket.org/s_l_teichmann/i3stuff/cmd/i3switch

  Manual build:

  $ go get -u github.com/proxypoke/i3ipc
  $ hg clone https://s_l_teichmann@bitbucket.org/s_l_teichmann/i3stuff
  $ cd i3stuff/cmd/i3switch
  $ go build

  Place the resulting i3win somewhere in your path.
  Move the resulting i3switch binary into your PATH. If you want
  to use the move feature (hard) link it to i3move and i3follow, e.g.:

  $ cp i3switch ~/bin
  $ cd ~/bin
  $ ln i3switch i3move
  $ ln i3switch i3follow

Usage:

Put something like this into your .i3/config:

bindsym $mod+t               exec i3switch $(dmenu -p "workspace:" 0>&-)
bindsym $mod+Shift+t         exec i3move   $(dmenu -p "move to workspace:" 0>&-)
bindsym $mod+Control+Shift+t exec i3follow $(dmenu -p "follow to workspace:" 0>&-)

