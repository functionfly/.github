/**
 * useNewFunction — convenience re-export of useFunctionEditor for the /functions/new route.
 *
 * The full implementation lives in useFunctionEditor.ts.
 * This file exists as a named alias for discoverability.
 */
export { useFunctionEditor as useNewFunction } from '@/pages/FunctionsPage/FunctionEditorPage/useFunctionEditor';
export type { FunctionEditorModel as NewFunctionModel } from '@/pages/FunctionsPage/FunctionEditorPage/useFunctionEditor';
