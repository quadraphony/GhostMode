package navigator

import (
	"errors"

	"ghost-browser/pkg/types"
)

var (
	ErrNoCurrentPage = errors.New("no current page")
	ErrNoBackHistory = errors.New("no back history")
	ErrNoForward     = errors.New("no forward history")
)

type Navigator struct {
	state types.NavigationState
}

func New(entries []types.HistoryEntry) *Navigator {
	return &Navigator{state: types.NavigationState{HistoryEntries: append([]types.HistoryEntry(nil), entries...)}}
}

func (n *Navigator) Current() *types.Page {
	return n.state.CurrentPage
}

func (n *Navigator) History() []types.HistoryEntry {
	return append([]types.HistoryEntry(nil), n.state.HistoryEntries...)
}

func (n *Navigator) Push(page *types.Page, entry types.HistoryEntry) {
	if n.state.CurrentPage != nil {
		n.state.BackStack = append(n.state.BackStack, n.state.CurrentPage)
	}
	n.state.CurrentPage = page
	n.state.ForwardStack = nil
	n.state.HistoryEntries = append(n.state.HistoryEntries, entry)
}

func (n *Navigator) ReplaceCurrent(page *types.Page) {
	n.state.CurrentPage = page
}

func (n *Navigator) Back() (*types.Page, error) {
	if len(n.state.BackStack) == 0 {
		return nil, ErrNoBackHistory
	}
	if n.state.CurrentPage != nil {
		n.state.ForwardStack = append(n.state.ForwardStack, n.state.CurrentPage)
	}
	last := n.state.BackStack[len(n.state.BackStack)-1]
	n.state.BackStack = n.state.BackStack[:len(n.state.BackStack)-1]
	n.state.CurrentPage = last
	return last, nil
}

func (n *Navigator) Forward() (*types.Page, error) {
	if len(n.state.ForwardStack) == 0 {
		return nil, ErrNoForward
	}
	if n.state.CurrentPage != nil {
		n.state.BackStack = append(n.state.BackStack, n.state.CurrentPage)
	}
	last := n.state.ForwardStack[len(n.state.ForwardStack)-1]
	n.state.ForwardStack = n.state.ForwardStack[:len(n.state.ForwardStack)-1]
	n.state.CurrentPage = last
	return last, nil
}
