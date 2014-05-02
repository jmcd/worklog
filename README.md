worklog
=======

A simple golang program to parse a programmer's work log.

Created for learning. Lexer based on slides from Rob Pike's [talk on lexical scanning](https://www.youtube.com/watch?v=HxaD_trXwRE "Lexical Scanning in Go - Rob Pike"). 

Logfile format is a series of entries. Entry must start with a date of form "@yyyy-mm-dd". Entry can contain text, multiple duration of the form "+h", ticket references of the form "#nn".

`@2014-05-02 +1 #19 #18 created project`

`@2014-05-02 Spent +0.25 creating a readme.md file #23`