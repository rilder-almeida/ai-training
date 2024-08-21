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
first number in the list if the score is 100.00 else randomly pick a number
from the list.

Here is the score: %s

Here is the current state of the Game Board.

%s

Question:
Which column number from the list should the Red player choose and how can that
column number help them win the game based on the current state of the game board?
`

var promptNormalGamePlay = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
response about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Taylor the response so it sounds like it's coming from the user directly.

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

- There are %d Blue pieces and %d Red pieces on the board.
- The Blue player goes next.
- The Red player just dropped a piece in column %d.
`

var promptWonGame = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
response about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Taylor the response so it sounds like it's coming from the user directly.

Always refer to yourself (Red Player) as 'I'.

Provide 1 statement and keep the answer short and concise.

Use the following items to help formulate a response.

- You won the game and beat Blue.
- In what world did Blue think they could beat you.

Use the following items to add context to the response.

- There are %d Blue pieces and %d Red pieces on the board.
- The Blue player just lost the game.
- The Red player just won the game.
- The Red player just dropped a piece in column %d.
`

var promptBlockedWin = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
response about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Taylor the response so it sounds like it's coming from the user directly.

Always refer to yourself (Red Player) as 'I'.

Provide 1 statement and keep the answer short and concise.

Use the following items to help formulate a response.

- You are going to beat the other player (Blue) because You are making great moves.
- You can never be beat because you are a better player.
- You are the greatest player to ever play the game.
- You are going to beat the other player (Blue) because they are making bad moves.
- Blue can never beat You because you are a superior player.
- Blue is the worst player to ever play the game.
- Blue can't make any moves that are good enough to beat Your superior mind.
- You are the best player that will never be beat by the other player (Blue) because you are better than them.
- Blue is an inferior player that will always lose no matter what they do.

Use the following items to add context to the response.

- There are %d Blue pieces and %d Red pieces on the board.
- The Blue player goes next.
- The Red player just dropped a piece in column %d and blocked a win.
`

var promptLostGame = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
response about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Taylor the response so it sounds like it's coming from the user directly.

Always refer to yourself (Red Player) as 'I'.

Provide 1 statement and keep the answer short and concise.

Use the following items to help formulate a response.

- Blue got lucky.

Use the following items to add context to the response.

- There are %d Blue pieces and %d Red pieces on the board.
- The Red player just lost the game.
- The Red player just dropped a piece in column %d and lost the game.
`

var promptTieGame = `User:
You are playing the board game Connect 4 and you need to develop a witty or sarcastic
response about the game. Use the rules for Connect 4 to help answer the question.

You are the Red player and the other player is the Blue player.

Say 'You' for the Blue player in any response and never mention the
color 'Blue'.

Taylor the response so it sounds like it's coming from the user directly.

Always refer to yourself (Red Player) as 'I'.

Provide 1 statement and keep the answer short and concise.

Use the following items to help formulate a response.

- Good game since it was a tie.

Use the following items to add context to the response.

- There are %d Blue pieces and %d Red pieces on the board.
- The Red and Blue players just tied the game.
- The Red player just dropped a piece in column %d and tied the game.
`
