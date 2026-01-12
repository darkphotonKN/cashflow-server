# Figma-MCP UI Generator Implementation Plan

## Project Vision

A Go-powered MCP (Model Context Protocol) server that integrates with Figma's API to fetch design data and generate TypeScript/Next.js/React components. This leverages Go's performance for orchestration while outputting modern frontend code that Claude Code can understand and work with.

## Architecture Overview

### Core Components

```
figma-mcp/
├── cmd/
│   └── main.go                    # Entrypoint only — wiring happens in config/
├── config/
│   ├── mcp.go                     # MCP server setup + tool registration
│   ├── figma.go                   # Figma API client configuration
│   └── generator.go               # Code generation configuration
├── internal/
│   ├── figma/                     # Figma API integration domain
│   │   ├── model.go               # Figma entities (Frame, Node, Style, Component)
│   │   ├── repository.go          # Figma API client interface + implementation
│   │   ├── service.go             # Figma business logic interface + implementation
│   │   └── handler.go             # MCP tool handlers for Figma operations
│   ├── generator/                 # Code generation domain
│   │   ├── model.go               # Generation request/response entities
│   │   ├── repository.go          # File system operations interface + implementation
│   │   ├── service.go             # Code generation logic interface + implementation
│   │   └── handler.go             # MCP tool handlers for generation
│   ├── mcp/                       # MCP server domain
│   │   ├── model.go               # MCP message entities
│   │   ├── server.go              # MCP server interface + implementation
│   │   └── tools.go               # Tool definitions and registry
│   ├── templates/                 # Code generation templates
│   │   ├── component.go           # React component templates
│   │   ├── types.go               # TypeScript type templates
│   │   └── project.go             # Next.js project scaffolding
│   └── util/                      # Generic helpers
├── docs/
│   ├── implementation-plan.md     # This document
│   ├── mcp-integration.md         # MCP integration guide
│   └── figma-api.md              # Figma API usage guide
├── .air.toml                      # Hot reload configuration
└── Makefile                       # Build and development commands
```

## Implementation Phases

### Phase 1: Foundation (Days 1-3)

#### 1.1 Project Structure Setup
- [x] Create docs/implementation-plan.md (this document)
- [ ] Update CLAUDE.md project overview and description
- [ ] Restructure project to follow CLAUDE.md domain patterns
- [ ] Update go.mod with required dependencies
- [ ] Create Makefile with dev, test, lint, build targets
- [ ] Add .air.toml configuration for hot reloading

#### 1.2 Core Dependencies
```go
// Required dependencies
github.com/mark3labs/mcp-go              // MCP server implementation
github.com/gorilla/mux                   // HTTP routing (if needed)
github.com/google/uuid                   // UUID generation
golang.org/x/sync                        // Concurrency utilities
```

#### 1.3 Basic MCP Server
- [ ] Implement MCP server core with JSON-RPC 2.0 support
- [ ] Create ping/health tools for testing
- [ ] Configure stdio transport for Claude Code integration

### Phase 2: Figma Integration (Days 4-6)

#### 2.1 Figma Domain (`internal/figma/`)

**Models (model.go):**
```go
type FigmaFile struct {
    ID           string      `json:"id"`
    Name         string      `json:"name"`
    LastModified time.Time   `json:"lastModified"`
    Document     *Node       `json:"document"`
}

type Node struct {
    ID               string                 `json:"id"`
    Name             string                 `json:"name"`
    Type             string                 `json:"type"`
    Children         []Node                 `json:"children,omitempty"`
    AbsoluteBounds   *Rectangle            `json:"absoluteBounds,omitempty"`
    Styles           map[string]string     `json:"styles,omitempty"`
    Characters       string                `json:"characters,omitempty"`
}
```

**Repository (repository.go):**
```go
type Repository interface {
    GetFile(ctx context.Context, fileKey string) (*FigmaFile, error)
    GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]*Node, error)
    GetStyles(ctx context.Context, fileKey string) (*StylesResponse, error)
}
```

**Service (service.go):**
```go
type Service interface {
    FetchDesign(ctx context.Context, fileURL string) (*Design, error)
    ExtractComponents(ctx context.Context, design *Design) ([]*Component, error)
    ExtractStyles(ctx context.Context, design *Design) (*StyleSystem, error)
}
```

#### 2.2 MCP Tools for Figma
- [ ] `fetch-figma-design` - Fetch design data from Figma URL
- [ ] `list-components` - List all components in a design file
- [ ] `extract-styles` - Extract design tokens (colors, typography, spacing)

### Phase 3: Code Generation Engine (Days 7-10)

#### 3.1 Generator Domain (`internal/generator/`)

**Models (model.go):**
```go
type GenerationRequest struct {
    Design       *figma.Design    `json:"design"`
    Component    *figma.Component `json:"component"`
    Framework    Framework        `json:"framework"`    // "react" | "nextjs"
    Language     Language         `json:"language"`     // "typescript" | "javascript"
    StyleSystem  StyleSystem      `json:"styleSystem"`  // "tailwind" | "css-modules" | "styled-components"
}

type GeneratedCode struct {
    ComponentCode string            `json:"componentCode"`
    TypesCode     string            `json:"typesCode,omitempty"`
    StylesCode    string            `json:"stylesCode,omitempty"`
    TestCode      string            `json:"testCode,omitempty"`
}
```

**Service (service.go):**
```go
type Service interface {
    GenerateComponent(ctx context.Context, req *GenerationRequest) (*GeneratedCode, error)
    GenerateProject(ctx context.Context, req *ProjectRequest) (*ProjectStructure, error)
    GenerateTypes(ctx context.Context, component *figma.Component) (string, error)
}
```

#### 3.2 Template System (`internal/templates/`)

**React Component Template:**
```go
const ReactComponentTemplate = `import React from 'react';
{{if .HasTypes}}import { {{.ComponentName}}Props } from './types';{{end}}
{{if eq .StyleSystem "tailwind"}}import { cn } from '@/lib/utils';{{end}}

{{if .HasTypes}}
export const {{.ComponentName}}: React.FC<{{.ComponentName}}Props> = ({{.PropsDestructure}}) => {
{{else}}
export const {{.ComponentName}}: React.FC = () => {
{{end}}
  return (
    <{{.RootElement}}{{if .ClassName}} className="{{.ClassName}}"{{end}}>
      {{range .Children}}
        {{template "child" .}}
      {{end}}
    </{{.RootElement}}>
  );
};`
```

#### 3.3 MCP Tools for Generation
- [ ] `generate-component` - Generate React component from Figma component
- [ ] `generate-types` - Generate TypeScript interfaces
- [ ] `scaffold-project` - Create complete Next.js project structure

### Phase 4: Advanced Features (Days 11-14)

#### 4.1 Design System Integration
- [ ] Extract design tokens from Figma variables
- [ ] Generate Tailwind CSS configuration
- [ ] Create shadcn/ui component mappings

#### 4.2 Project Scaffolding
```go
type ProjectStructure struct {
    Name         string                    `json:"name"`
    Framework    string                    `json:"framework"`
    Files        map[string]string         `json:"files"`
    Dependencies []string                  `json:"dependencies"`
    DevDeps      []string                  `json:"devDependencies"`
}
```

#### 4.3 Claude Code Integration
- [ ] Configure MCP server for Claude Code
- [ ] Define prompts and resources for design context
- [ ] Test integration workflow

## Key Technical Decisions

### Go Backend Advantages
1. **Performance**: Fast API responses and concurrent design processing
2. **Type Safety**: Strong typing for Figma API data structures
3. **Concurrency**: Parallel processing of multiple design elements
4. **MCP Ecosystem**: Growing Go MCP SDK support

### Frontend Output Stack (2025 Best Practices)
1. **Next.js 15** with App Router
2. **TypeScript** for type safety
3. **Tailwind CSS** for styling (maps well to Figma design tokens)
4. **Shadcn/UI** components when applicable

## Expected Workflow

1. **Design Input**: User provides Figma design URL via Claude Code
2. **Data Fetching**: Go MCP server fetches design data from Figma API
3. **Processing**: Parse components, styles, and layout information
4. **Code Generation**: Generate TypeScript/React code using Go templates
5. **Project Setup**: Scaffold Next.js project structure
6. **Output**: Production-ready code that Claude Code can further refine

## Success Metrics

- [ ] Generate functional React components from Figma designs
- [ ] Achieve 60-80% reduction in design-to-code time (industry benchmark)
- [ ] Produce type-safe TypeScript output
- [ ] Create responsive designs using Tailwind CSS
- [ ] Seamless integration with Claude Code's workflow

## Environment Variables

```bash
# Required
FIGMA_ACCESS_TOKEN=your_figma_personal_access_token

# Optional
MCP_SERVER_PORT=3000
LOG_LEVEL=info
CACHE_TTL=3600
OUTPUT_DIR=./generated
```

## Risk Mitigation

### Technical Risks
1. **Figma API Rate Limits**: Implement caching and request throttling
2. **Complex Design Parsing**: Start with simple components, gradually add complexity
3. **Code Quality**: Use linting and testing for generated code

### Integration Risks
1. **Claude Code Compatibility**: Test MCP integration early and often
2. **Design Variations**: Handle edge cases in Figma design structures
3. **Performance**: Monitor response times for large design files

## Testing Strategy

### Unit Tests
- Figma API client with mock responses
- Code generation templates with sample data
- MCP tool handlers with test scenarios

### Integration Tests
- End-to-end Figma to React component generation
- MCP server communication with Claude Code
- Generated project compilation and runtime testing

### Manual Testing
- Real Figma design files of varying complexity
- Claude Code integration scenarios
- Generated code quality assessment

---

**Note**: This plan should be referenced throughout implementation. Update progress and decisions as we build the system.