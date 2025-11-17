package domain

// GitHubRepoInfo represents information about a GitHub repository.
type GitHubRepoInfo struct {
	Owner       string
	Name        string
	FullName    string // owner/repo
	Description string
	Stars       int
	Forks       int
	Issues      int
	IsPrivate   bool
	URL         string
	HTMLURL     string
	DefaultBranch string
}

// RepoPath returns the owner/repo path.
func (g *GitHubRepoInfo) RepoPath() string {
	if g.FullName != "" {
		return g.FullName
	}
	return g.Owner + "/" + g.Name
}
