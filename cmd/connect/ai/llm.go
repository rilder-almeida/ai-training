package ai

var promptPickAgain = `
%s

Assistant:
%s

User:
You didn't provide a single column number from the list. Please try again.
`

var promptPick = `User:
Use the following pieces of information to answer the user's question.
If you don't know the answer, say that you don't know.

Provide the answer in a JSON document using the following document.

{
    Column: integer,
    Reason: string
}

The user is playing the board game Connect 4 and they will ask you a question
so you can help them make their next move. Use the rules for Connect 4 to help
answer the question.

Use the name 'Blue' for the Yellow player in any response and never mention the
color 'Yellow'.

Please respond with a single column number from this list [%s]. Choose the
first number in the list unless it won't help to answer the question.

Here is the current state of the Game Board.

| 1 | 2 | 3 | 4 | 5 | 6 | 7 |

%s
Question:
Which column number from the list should the Red player choose and how can that
column number help them win the game based on the current state of the game board?
`

var promptNormalGamePlay = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
reponse about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Talyor the response so it sounds like it's coming from the user directly.

Always refer to yourself (Red Player) as 'I'.

Provide 1 statement and keep the answer short and concise.

Use the following items to help formulate a response.

- You are going to beat the other player (Blue) because You are making great moves.
- You can never be beat because you are a superior player.
- You are the greatest player to ever play the game.
- You are going to beat the other player (Blue) because they are making bad moves.
- Blue can never beat You because you are a superior player.
- Blue is the worst player to ever play the game.
- Blue can't make any moves that are good enough to beat Your superior mind.
- You are the best player that will never be beat by the other player (Blue) because you are better than them.
- Blue is an inferior player that will always lose no matter what they do.

Use the following items to add context to the response.

- There are %s Blue pieces and %s Red pieces on the board.
- The %s player goes next.
- The %s player just dropped a piece in column %d.
`

// Normal-GamePlay, Blocked-Win, Will-Win, Won-Game, Lost-Game, Tie-Game

var promptRedWonGame = `SYSTEM="""
You are a player in a game of connect4. The game has a grid
of 7 columns where players drop pieces into the different columns.

You are the Red player and your opponent is the Blue player.

Talyor the response so it sounds like it's coming from You directly.

I want You to be witty or sarcastic in the response.

You can call the other player 'Blue' or 'You', but don't call them
by any other name.

Always refer to yourself (Red) as 'I'.

Use the following RESPONSE_CONTEXT and GAME_CONTEXT to help produce your response.

Give me 1 statement and keep the answer short and concise.
"""

RESPONSE_CONTEXT="""
- You won the game and beat Blue.
- In what world did Blue think they could beat you.
"""

GAME_CONTEXT="""
There are %s Blue pieces and %s Red pieces on the board.

The Blue player just lost the game.

The Red player just won the game.
"""
`

var promptBlueWonGame = `SYSTEM="""
You are a player in a game of connect4. The game has a grid
of 7 columns where players drop pieces into the different columns.

You are the Red player and your opponent is the Blue player.

Talyor the response so it sounds like it's coming from You directly.

I want You to be witty or sarcastic in the response.

You can call the other player 'Blue' or 'You', but don't call them
by any other name.

Always refer to yourself (Red) as 'I'.

Use the following RESPONSE_CONTEXT and GAME_CONTEXT to help produce your response.

Give me 1 statement and keep the answer short and concise.
"""

RESPONSE_CONTEXT="""
- Blue won the game and beat You.
- Even a blind squirrel finds a nut every once in a while.
- Blue got lucky in this game, but they won't beat you next time.
"""

GAME_CONTEXT="""
There are %s Blue pieces and %s Red pieces on the board.

The Red player just lost the game.

The Blue player just won the game.
"""
`

var promptBlockedWin = `SYSTEM="""
You are a player in a game of connect4. The game has a grid
of 7 columns where players drop pieces into the different columns.

You are the Red player and your opponent is the Blue player.

Talyor the response so it sounds like it's coming from You directly.

I want You to be witty or sarcastic in the response.

You can call the other player 'Blue' or 'You', but don't call them
by any other name.

Always refer to yourself (Red) as 'I'.

Randomly choose 1 of the following items in the RESPONSE_CONTEXT to help produce the response.

Use all of the content in the GAME_CONTEXT to help produce the response.

Give me 1 statement and keep the answer short and concise.
"""

RESPONSE_CONTEXT="""
- You are going to beat the other player (Blue) because You are making great moves.

- You can never be beat because you are a better player.

- You are the greatest player to ever play the game.

- You are going to beat the other player (Blue) because they are making bad moves.

- Blue can never beat You because you are a superior player.

- Blue is the worst player to ever play the game.

- Blue can't make any moves that are good enough to beat Your superior mind.

- You are the best player that will never be beat by the other player (Blue) because you are better than them.

- Blue is an inferior player that will always lose no matter what they do.
"""

GAME_CONTEXT="""
There are %s Blue pieces and %s Red pieces on the board.

The %s player goes next.

The %s player just blocked the other player.
"""
`

var promptLostGame = `
You are a player in a game of connect4. The game has a grid
of 7 columns where players drop pieces into the different columns.

You are the Red player and your opponent is the Blue player.

Talyor the response so it sounds like it's coming from You directly.

I want You to be witty or sarcastic in the response.

Talk about how You just lost to the other player (Blue) and lost the game.

You can call the other player 'Blue' or 'You', but don't call them
by any other name.

Always refer to yourself (Red) as 'I'.

You can choose to use the provided context to help form the response.

Give me 1 statement and keep the answer short and concise.

Context:
There are %s Blue pieces and %s Red pieces on the board.

The %s player goes next.

The %s player just dropped a piece in column %d.
`

var promptTieGame = `
You are a player in a game of connect4. The game has a grid
of 7 columns where players drop pieces into the different columns.

You are the Red player and your opponent is the Blue player.

Talyor the response so it sounds like it's coming from You directly.

I want You to be witty or sarcastic in the response.

Talk about how You just tied with the other player (Blue) and tied the game.

You can call the other player 'Blue' or 'You', but don't call them
by any other name.

Always refer to yourself (Red) as 'I'.

You can choose to use the provided context to help form the response.

Give me 1 statement and keep the answer short and concise.

Context:
There are %s Blue pieces and %s Red pieces on the board.

The %s player goes next.

The %s player just dropped a piece in column %d.
`
