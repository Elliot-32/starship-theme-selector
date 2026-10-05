Output docs/demo.gif

Require ssts
Require starship

Set Shell "bash"
Set FontFamily "FiraCode Nerd Font"
Set FontSize 22
Set Width 1200
Set Height 720
Set Padding 18
Set Framerate 30
Set TypingSpeed 45ms

Type "ssts"
Enter
Wait+Screen@10s /Starship Theme Selector/
Sleep 1s

Down@250ms 4
Sleep 800ms
Up@250ms 2
Sleep 800ms

Type "/"
Type@90ms "tokyo-night"
Wait+Screen@10s /tokyo-night/
Sleep 1s
Enter
Sleep 500ms
Enter
Sleep 2s
