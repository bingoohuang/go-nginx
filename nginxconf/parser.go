package nginxconf

import (
	"container/list"
	"sync"

	"github.com/pkg/errors"
)

// NginxConfigureBlock represent a block in nginx configure file.
// The content of a nginx configure file should be a block.
type NginxConfigureBlock []NginxConfigureCommand

// NginxConfigureCommand represenct a command in nginx configure file.
type NginxConfigureCommand struct {
	// Words compose the command
	Words []string

	// Block follow the command
	Block NginxConfigureBlock
}

type parser struct {
	sync.Mutex
	*Scanner
}

// nolint:gochecknoglobals
var (
	emptyBlock   = NginxConfigureBlock(nil)
	emptyCommand = NginxConfigureCommand{}
)

// Parse the content of nginx configure file into NginxConfigureBlock.
func Parse(content []byte) (blk NginxConfigureBlock, err error) {
	var p parser
	return p.parse(content)
}

func (p *parser) parse(content []byte) (blk NginxConfigureBlock, err error) {
	p.Lock()
	defer p.Unlock()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	p.Scanner = NewScanner(content)
	cmds := list.New()

ForLoop:
	for {
		token := p.Scan()
		switch token.Typ {
		case EOF:
			break ForLoop
		case Word:
			cmd, err := p.scanCommand(token.Lit)
			if err != nil {
				return nil, err
			}
			cmds.PushBack(cmd)
		case Comment:
			continue
		default:
			return nil, errors.Wrapf(ErrSyntax, "unexpected global token %s at line %d", token.Typ, p.line)
		}
	}

	cfg := make([]NginxConfigureCommand, cmds.Len())

	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		cfg[i] = cmd.Value.(NginxConfigureCommand)
	}

	return cfg, nil
}

func (p *parser) scanCommand(startWord string) (NginxConfigureCommand, error) {
	words := list.New()

	if startWord != "" {
		words.PushBack(startWord)
	}

	var (
		err   error
		block NginxConfigureBlock
	)

ForLoop:
	for {
		token := p.Scan()
		switch token.Typ {
		case EOF:
			return emptyCommand, errors.Wrapf(ErrSyntax, "missing terminating token at line %d", p.line)
		case braceOpen:
			block, err = p.scanBlock()
			if err != nil {
				return emptyCommand, err
			}
			break ForLoop
		case semicolon:
			break ForLoop
		case Comment:
			continue
		case Word:
			words.PushBack(token.Lit)
		default:
			return emptyCommand, errors.Wrapf(ErrSyntax, "unexpected command token %s at line %d", token.Typ, p.line)
		}
	}

	cmd := NginxConfigureCommand{
		Words: make([]string, words.Len()),
		Block: block,
	}

	for i, word := 0, words.Front(); word != nil; i, word = i+1, word.Next() {
		cmd.Words[i] = word.Value.(string)
	}

	return cmd, nil
}

func (p *parser) scanBlock() (NginxConfigureBlock, error) {
	cmds := list.New()
ForLoop:
	for {
		token := p.Scan()
		switch token.Typ {
		case EOF:
			return emptyBlock, errors.Wrapf(ErrSyntax, "missing terminating token at line %d", p.line)
		case braceClose:
			break ForLoop
		case Comment:
			continue
		case Word:
			cmd, err := p.scanCommand(token.Lit)
			if err != nil {
				return emptyBlock, err
			}
			cmds.PushBack(cmd)
		default:
			return emptyBlock, errors.Wrapf(ErrSyntax, "unexpected block token %s at line %d", token.Typ, p.line)
		}
	}

	block := make([]NginxConfigureCommand, cmds.Len())

	for i, cmd := 0, cmds.Front(); cmd != nil; i, cmd = i+1, cmd.Next() {
		block[i] = cmd.Value.(NginxConfigureCommand)
	}

	return block, nil
}
