/**
 * useCodeIntelligence Hook
 * Main hook and specialized hooks for code intelligence components
 */

import { useCallback, useMemo } from 'react'
import { useCodeIntelligenceStore } from '../stores/codeIntelligenceStore'

// Main hook returning the full store
export const useCodeIntelligence = () => useCodeIntelligenceStore()

// Editor hook
export const useCodeEditor = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    value: store.editor.value,
    language: store.editor.language,
    filePath: store.editor.filePath,
    cursorPosition: store.editor.cursorPosition,
    selection: store.editor.selection,
    symbols: store.editor.symbols,
    diagnostics: store.editor.diagnostics,
    setValue: store.setEditorValue,
    setLanguage: store.setEditorLanguage,
    setFilePath: store.setEditorFilePath,
    setCursorPosition: store.setEditorCursorPosition,
    setSelection: store.setEditorSelection,
    setSymbols: store.setEditorSymbols,
    setDiagnostics: store.setEditorDiagnostics,
  }), [store])
}

// AST Explorer hook
export const useASTExplorer = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    ast: store.ast,
    selectedNodeId: store.selectedASTNodeId,
    expandedNodes: store.expandedASTNodes,
    searchQuery: store.astSearchQuery,
    setAST: store.setAST,
    selectNode: store.selectASTNode,
    toggleNode: store.toggleASTNode,
    setSearchQuery: store.setASTSearchQuery,
  }), [store])
}

// Dependencies hook
export const useDependencies = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    nodes: store.dependencies.nodes,
    edges: store.dependencies.edges,
    selectedNodeId: store.selectedDependencyNodeId,
    colorMetric: store.dependencyColorMetric,
    setDependencies: store.setDependencies,
    selectNode: store.selectDependencyNode,
    setColorMetric: store.setDependencyColorMetric,
  }), [store])
}

// Diff Viewer hook
export const useDiffViewer = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    files: store.diffFiles,
    selectedFileId: store.selectedDiffFileId,
    setDiffFiles: store.setDiffFiles,
    selectFile: store.selectDiffFile,
  }), [store])
}

// Architecture Map hook
export const useArchitectureMap = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    nodes: store.architecture.nodes,
    connections: store.architecture.connections,
    selectedNodeId: store.selectedArchitectureNodeId,
    expandedNodes: store.expandedArchitectureNodes,
    layout: store.architectureLayout,
    showMetrics: store.showArchitectureMetrics,
    setArchitecture: store.setArchitecture,
    selectNode: store.selectArchitectureNode,
    toggleNode: store.toggleArchitectureNode,
    setLayout: store.setArchitectureLayout,
    setShowMetrics: store.setShowArchitectureMetrics,
  }), [store])
}

// Smart Refactor hook
export const useSmartRefactor = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    opportunities: store.refactorOpportunities,
    selectedOpportunityId: store.selectedRefactorOpportunityId,
    setOpportunities: store.setRefactorOpportunities,
    selectOpportunity: store.selectRefactorOpportunity,
  }), [store])
}

// Code Generation hook
export const useCodeGeneration = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    generatedCode: store.generatedCode,
    isGenerating: store.isGenerating,
    error: store.generationError,
    showExplanation: store.showGenerationExplanation,
    setGeneratedCode: store.setGeneratedCode,
    setIsGenerating: store.setIsGenerating,
    setError: store.setGenerationError,
    toggleExplanation: store.toggleGenerationExplanation,
    clear: store.clearGeneratedCode,
  }), [store])
}

// Inline AI hook
export const useInlineAI = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    suggestions: store.inlineSuggestions,
    currentSuggestion: store.currentInlineSuggestion,
    enabled: store.inlineAIEnabled,
    setSuggestions: store.setInlineSuggestions,
    setCurrentSuggestion: store.setCurrentInlineSuggestion,
    toggle: store.toggleInlineAI,
  }), [store])
}

// Intent Explorer hook
export const useIntentExplorer = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    intents: store.intents,
    selectedIntentId: store.selectedIntentId,
    showReasoning: store.showIntentReasoning,
    setIntents: store.setIntents,
    selectIntent: store.selectIntent,
    toggleReasoning: store.toggleIntentReasoning,
  }), [store])
}

// Semantic Search hook
export const useSemanticSearch = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    query: store.searchQuery,
    results: store.searchResults,
    selectedResultId: store.selectedSearchResultId,
    searchType: store.searchType,
    isSearching: store.isSearching,
    setQuery: store.setSearchQuery,
    setResults: store.setSearchResults,
    selectResult: store.selectSearchResult,
    setSearchType: store.setSearchType,
    setIsSearching: store.setIsSearching,
    clearResults: store.clearSearchResults,
  }), [store])
}

// Lineage hook
export const useCodeLineage = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    nodes: store.lineageNodes,
    selectedNodeId: store.selectedLineageNodeId,
    focusedFilePath: store.focusedFilePath,
    setNodes: store.setLineageNodes,
    selectNode: store.selectLineageNode,
    setFocusedFilePath: store.setFocusedFilePath,
  }), [store])
}

// Risk Analyzer hook
export const useRiskAnalyzer = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    risks: store.riskIndicators,
    selectedRiskId: store.selectedRiskId,
    showMetrics: store.showRiskMetrics,
    setRisks: store.setRiskIndicators,
    selectRisk: store.selectRisk,
    setShowMetrics: store.setShowRiskMetrics,
  }), [store])
}

// Import Graph hook
export const useImportGraph = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    imports: store.imports,
    edges: store.importEdges,
    selectedNodeId: store.selectedImportNodeId,
    filePath: store.selectedImportFilePath,
    setImports: store.setImports,
    selectNode: store.selectImportNode,
    setFilePath: store.setSelectedImportFilePath,
  }), [store])
}

// Execution Editor hook
export const useExecutionEditor = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    executionPoints: store.executionPoints,
    currentPointId: store.currentExecutionPointId,
    breakpoints: store.breakpoints,
    watchExpressions: store.watchExpressions,
    isRunning: store.isRunning,
    setExecutionPoints: store.setExecutionPoints,
    selectPoint: store.selectExecutionPoint,
    toggleBreakpoint: store.toggleBreakpoint,
    addWatch: store.addWatchExpression,
    removeWatch: store.removeWatchExpression,
    setIsRunning: store.setIsRunning,
  }), [store])
}

// AI Completion Inspector hook
export const useCompletionInspector = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    completions: store.completions,
    selectedCompletionId: store.selectedCompletionId,
    currentCompletion: store.currentCompletion,
    setCompletions: store.setCompletions,
    selectCompletion: store.selectCompletion,
  }), [store])
}

// Simulation hook
export const useRefactorSimulation = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    simulation: store.simulation,
    step: store.simulationStep,
    setSimulation: store.setSimulation,
    stepForward: store.stepSimulationForward,
    stepBackward: store.stepSimulationBackward,
    clear: store.clearSimulation,
  }), [store])
}

// Constraints hook
export const useArchitectureConstraints = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    constraints: store.constraints,
    selectedConstraintId: store.selectedConstraintId,
    setConstraints: store.setConstraints,
    selectConstraint: store.selectConstraint,
  }), [store])
}

// Ownership hook
export const useCodeOwnership = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    ownerships: store.ownerships,
    selectedFilePath: store.selectedFilePath,
    selectedOwnerId: store.selectedOwnerId,
    setOwnerships: store.setOwnerships,
    selectFile: store.selectFileOwnership,
    selectOwner: store.selectOwner,
  }), [store])
}

// UI State hook
export const useCodeIntelligenceUI = () => {
  const store = useCodeIntelligenceStore()
  
  return useMemo(() => ({
    activePanel: store.activePanel,
    sidebarCollapsed: store.sidebarCollapsed,
    setActivePanel: store.setActivePanel,
    toggleSidebar: store.toggleSidebar,
  }), [store])
}
