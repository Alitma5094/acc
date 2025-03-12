from enum import Enum, auto
from os import curdir
from typing import Any
class TokenType(Enum):
    # Single-character tokens
    OPEN_PAREN = auto()
    CLOSE_PAREN = auto()
    OPEN_BRACE = auto()
    CLOSE_BRACE = auto()
    SEMICOLON = auto()

    # Literals
    IDENTIFIER = auto()
    CONSTANT = auto()

    # Keywords
    INT = auto()
    VOID = auto()
    RETURN = auto()

keywords = {"int": TokenType.INT, "void": TokenType.VOID, "return": TokenType.RETURN}

class Token:
    def __init__(self, type: TokenType, lexeme: str, literal: Any, line: int) -> None:
        self.type = type
        self.lexeme = lexeme
        self.literal = literal
        self.line = line

    def __str__(self) -> str:
        return f"{self.type.name} {self.lexeme} {self.literal}"

class Lexer:
    def __init__(self, source: str) -> None:
        self.source: str = source
        self.tokens = []

        self.start = 0
        self.current = 0
        self.line = 1

    def isAtEnd(self):
        return self.current >= len(self.source)

    def scanTokens(self):
        while not self.isAtEnd():
            self.start = self.current
            self.scanToken()
            
        return self.tokens
    
    def scanToken(self):
        c = self.advance()
        match c:
            case "(":
                self.add_token(TokenType.OPEN_PAREN, None)
            case ")":
                self.add_token(TokenType.CLOSE_PAREN, None)
            case "{":
                self.add_token(TokenType.OPEN_BRACE, None)
            case "}":
                self.add_token(TokenType.CLOSE_BRACE, None)
            case ";":
                self.add_token(TokenType.SEMICOLON, None)

            # Ignore whitespace
            case " ":
                pass
            case "\t":
                pass
            case "\r":
                pass

            case "\n":
                self.line += 1

            case _:
                if c.isdigit():
                    self.number()
                elif c.isalnum():
                    self.identifier()
                else:
                    raise Exception("Invalid char")

    def advance(self):
        c = self.source[self.current]
        self.current += 1
        return c

    def peek(self) -> str:
        if self.isAtEnd(): 
            return '\0'
        return self.source[self.current]

    def add_token(self, type: TokenType, literal: Any):
        text = self.source[self.start:self.current]
        self.tokens.append(Token(type, text, literal, self.line))
    
    def number(self):
        while self.peek().isdigit():
            self.advance()
        
        if not self.isAtEnd() and self.peek().isalpha():
            raise Exception(f"Invalid numeric format at line {self.line}: unexpected '{self.peek()}' after number")
    
        self.add_token(TokenType.CONSTANT, int(self.source[self.start:self.current]))
    
    def identifier(self):
        while self.peek().isalnum():
            self.advance()

        text = self.source[self.start:self.current]
        type = keywords[text] if text in keywords.keys() else TokenType.IDENTIFIER
        self.add_token(type, None)

