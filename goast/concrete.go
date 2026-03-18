package goast

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// KindName strips the "Kind" prefix from ast.Kind.String().
func KindName(k ast.Kind) string {
	s := k.String()
	if strings.HasPrefix(s, "Kind") {
		return s[4:]
	}
	return s
}

// GetConcreteValue returns the concrete AST struct for a node via its As*() method.
// Returns nil for token/keyword nodes that don't have a dedicated struct.
func GetConcreteValue(node *ast.Node) any {
	switch node.Kind {
	// Identifiers & Names
	case ast.KindIdentifier:
		return node.AsIdentifier()
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier()
	case ast.KindQualifiedName:
		return node.AsQualifiedName()
	case ast.KindComputedPropertyName:
		return node.AsComputedPropertyName()

	// Literals
	case ast.KindStringLiteral:
		return node.AsStringLiteral()
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral()
	case ast.KindBigIntLiteral:
		return node.AsBigIntLiteral()
	case ast.KindRegularExpressionLiteral:
		return node.AsRegularExpressionLiteral()
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral()
	case ast.KindTemplateHead:
		return node.AsTemplateHead()
	case ast.KindTemplateMiddle:
		return node.AsTemplateMiddle()
	case ast.KindTemplateTail:
		return node.AsTemplateTail()
	case ast.KindJsxText:
		return node.AsJsxText()

	// Expressions
	case ast.KindPrefixUnaryExpression:
		return node.AsPrefixUnaryExpression()
	case ast.KindPostfixUnaryExpression:
		return node.AsPostfixUnaryExpression()
	case ast.KindParenthesizedExpression:
		return node.AsParenthesizedExpression()
	case ast.KindTypeAssertionExpression:
		return node.AsTypeAssertion()
	case ast.KindAsExpression:
		return node.AsAsExpression()
	case ast.KindSatisfiesExpression:
		return node.AsSatisfiesExpression()
	case ast.KindExpressionWithTypeArguments:
		return node.AsExpressionWithTypeArguments()
	case ast.KindNonNullExpression:
		return node.AsNonNullExpression()
	case ast.KindDeleteExpression:
		return node.AsDeleteExpression()
	case ast.KindTypeOfExpression:
		return node.AsTypeOfExpression()
	case ast.KindVoidExpression:
		return node.AsVoidExpression()
	case ast.KindAwaitExpression:
		return node.AsAwaitExpression()
	case ast.KindBinaryExpression:
		return node.AsBinaryExpression()
	case ast.KindConditionalExpression:
		return node.AsConditionalExpression()
	case ast.KindCallExpression:
		return node.AsCallExpression()
	case ast.KindNewExpression:
		return node.AsNewExpression()
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression()
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression()
	case ast.KindTaggedTemplateExpression:
		return node.AsTaggedTemplateExpression()
	case ast.KindTemplateExpression:
		return node.AsTemplateExpression()
	case ast.KindYieldExpression:
		return node.AsYieldExpression()
	case ast.KindSpreadElement:
		return node.AsSpreadElement()
	case ast.KindObjectLiteralExpression:
		return node.AsObjectLiteralExpression()
	case ast.KindArrayLiteralExpression:
		return node.AsArrayLiteralExpression()
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression()
	case ast.KindArrowFunction:
		return node.AsArrowFunction()
	case ast.KindClassExpression:
		return node.AsClassExpression()
	case ast.KindMetaProperty:
		return node.AsMetaProperty()
	case ast.KindTemplateSpan:
		return node.AsTemplateSpan()
	case ast.KindPartiallyEmittedExpression:
		return node.AsPartiallyEmittedExpression()
	case ast.KindSyntheticExpression:
		return node.AsSyntheticExpression()
	case ast.KindSyntheticReferenceExpression:
		return node.AsSyntheticReferenceExpression()

	// Statements
	case ast.KindBlock:
		return node.AsBlock()
	case ast.KindEmptyStatement:
		return node.AsEmptyStatement()
	case ast.KindVariableStatement:
		return node.AsVariableStatement()
	case ast.KindExpressionStatement:
		return node.AsExpressionStatement()
	case ast.KindIfStatement:
		return node.AsIfStatement()
	case ast.KindDoStatement:
		return node.AsDoStatement()
	case ast.KindWhileStatement:
		return node.AsWhileStatement()
	case ast.KindForStatement:
		return node.AsForStatement()
	case ast.KindForInStatement, ast.KindForOfStatement:
		return node.AsForInOrOfStatement()
	case ast.KindContinueStatement:
		return node.AsContinueStatement()
	case ast.KindBreakStatement:
		return node.AsBreakStatement()
	case ast.KindReturnStatement:
		return node.AsReturnStatement()
	case ast.KindWithStatement:
		return node.AsWithStatement()
	case ast.KindSwitchStatement:
		return node.AsSwitchStatement()
	case ast.KindLabeledStatement:
		return node.AsLabeledStatement()
	case ast.KindThrowStatement:
		return node.AsThrowStatement()
	case ast.KindTryStatement:
		return node.AsTryStatement()
	case ast.KindDebuggerStatement:
		return node.AsDebuggerStatement()
	case ast.KindNotEmittedStatement:
		return node.AsNotEmittedStatement()
	case ast.KindModuleBlock:
		return node.AsModuleBlock()

	// Declarations
	case ast.KindVariableDeclaration:
		return node.AsVariableDeclaration()
	case ast.KindVariableDeclarationList:
		return node.AsVariableDeclarationList()
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration()
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration()
	case ast.KindInterfaceDeclaration:
		return node.AsInterfaceDeclaration()
	case ast.KindTypeAliasDeclaration:
		return node.AsTypeAliasDeclaration()
	case ast.KindEnumDeclaration:
		return node.AsEnumDeclaration()
	case ast.KindModuleDeclaration:
		return node.AsModuleDeclaration()
	case ast.KindImportDeclaration:
		return node.AsImportDeclaration()
	case ast.KindImportClause:
		return node.AsImportClause()
	case ast.KindNamespaceImport:
		return node.AsNamespaceImport()
	case ast.KindNamedImports:
		return node.AsNamedImports()
	case ast.KindImportSpecifier:
		return node.AsImportSpecifier()
	case ast.KindImportEqualsDeclaration:
		return node.AsImportEqualsDeclaration()
	case ast.KindExportAssignment:
		return node.AsExportAssignment()
	case ast.KindExportDeclaration:
		return node.AsExportDeclaration()
	case ast.KindNamedExports:
		return node.AsNamedExports()
	case ast.KindNamespaceExport:
		return node.AsNamespaceExport()
	case ast.KindExportSpecifier:
		return node.AsExportSpecifier()
	case ast.KindMissingDeclaration:
		return node.AsMissingDeclaration()
	case ast.KindExternalModuleReference:
		return node.AsExternalModuleReference()
	case ast.KindImportAttribute:
		return node.AsImportAttribute()
	case ast.KindImportAttributes:
		return node.AsImportAttributes()
	case ast.KindNamespaceExportDeclaration:
		return node.AsNamespaceExportDeclaration()
	case ast.KindCommonJSExport:
		return node.AsCommonJSExport()

	// Class members
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration()
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration()
	case ast.KindConstructor:
		return node.AsConstructorDeclaration()
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration()
	case ast.KindSetAccessor:
		return node.AsSetAccessorDeclaration()
	case ast.KindClassStaticBlockDeclaration:
		return node.AsClassStaticBlockDeclaration()
	case ast.KindSemicolonClassElement:
		return node.AsSemicolonClassElement()

	// Type members
	case ast.KindPropertySignature:
		return node.AsPropertySignatureDeclaration()
	case ast.KindMethodSignature:
		return node.AsMethodSignatureDeclaration()
	case ast.KindCallSignature:
		return node.AsCallSignatureDeclaration()
	case ast.KindConstructSignature:
		return node.AsConstructSignatureDeclaration()
	case ast.KindIndexSignature:
		return node.AsIndexSignatureDeclaration()

	// Parameters & Binding
	case ast.KindParameter:
		return node.AsParameterDeclaration()
	case ast.KindTypeParameter:
		return node.AsTypeParameter()
	case ast.KindDecorator:
		return node.AsDecorator()
	case ast.KindBindingElement:
		return node.AsBindingElement()
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		return node.AsBindingPattern()

	// Types
	case ast.KindTypePredicate:
		return node.AsTypePredicateNode()
	case ast.KindTypeReference:
		return node.AsTypeReferenceNode()
	case ast.KindFunctionType:
		return node.AsFunctionTypeNode()
	case ast.KindConstructorType:
		return node.AsConstructorTypeNode()
	case ast.KindTypeQuery:
		return node.AsTypeQueryNode()
	case ast.KindTypeLiteral:
		return node.AsTypeLiteralNode()
	case ast.KindArrayType:
		return node.AsArrayTypeNode()
	case ast.KindTupleType:
		return node.AsTupleTypeNode()
	case ast.KindOptionalType:
		return node.AsOptionalTypeNode()
	case ast.KindRestType:
		return node.AsRestTypeNode()
	case ast.KindUnionType:
		return node.AsUnionTypeNode()
	case ast.KindIntersectionType:
		return node.AsIntersectionTypeNode()
	case ast.KindConditionalType:
		return node.AsConditionalTypeNode()
	case ast.KindInferType:
		return node.AsInferTypeNode()
	case ast.KindParenthesizedType:
		return node.AsParenthesizedTypeNode()
	case ast.KindThisType:
		return node.AsThisTypeNode()
	case ast.KindTypeOperator:
		return node.AsTypeOperatorNode()
	case ast.KindIndexedAccessType:
		return node.AsIndexedAccessTypeNode()
	case ast.KindMappedType:
		return node.AsMappedTypeNode()
	case ast.KindLiteralType:
		return node.AsLiteralTypeNode()
	case ast.KindNamedTupleMember:
		return node.AsNamedTupleMember()
	case ast.KindTemplateLiteralType:
		return node.AsTemplateLiteralTypeNode()
	case ast.KindTemplateLiteralTypeSpan:
		return node.AsTemplateLiteralTypeSpan()
	case ast.KindImportType:
		return node.AsImportTypeNode()

	// Keyword type nodes
	case ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindNumberKeyword,
		ast.KindBigIntKeyword, ast.KindObjectKeyword, ast.KindBooleanKeyword,
		ast.KindStringKeyword, ast.KindSymbolKeyword, ast.KindVoidKeyword,
		ast.KindUndefinedKeyword, ast.KindNeverKeyword, ast.KindIntrinsicKeyword:
		return node.AsKeywordTypeNode()

	// Clauses
	case ast.KindCaseClause, ast.KindDefaultClause:
		return node.AsCaseOrDefaultClause()
	case ast.KindHeritageClause:
		return node.AsHeritageClause()
	case ast.KindCatchClause:
		return node.AsCatchClause()
	case ast.KindCaseBlock:
		return node.AsCaseBlock()

	// Object literal elements
	case ast.KindPropertyAssignment:
		return node.AsPropertyAssignment()
	case ast.KindShorthandPropertyAssignment:
		return node.AsShorthandPropertyAssignment()
	case ast.KindSpreadAssignment:
		return node.AsSpreadAssignment()
	case ast.KindEnumMember:
		return node.AsEnumMember()

	// JSX
	case ast.KindJsxElement:
		return node.AsJsxElement()
	case ast.KindJsxSelfClosingElement:
		return node.AsJsxSelfClosingElement()
	case ast.KindJsxOpeningElement:
		return node.AsJsxOpeningElement()
	case ast.KindJsxClosingElement:
		return node.AsJsxClosingElement()
	case ast.KindJsxFragment:
		return node.AsJsxFragment()
	case ast.KindJsxOpeningFragment:
		return node.AsJsxOpeningFragment()
	case ast.KindJsxClosingFragment:
		return node.AsJsxClosingFragment()
	case ast.KindJsxAttribute:
		return node.AsJsxAttribute()
	case ast.KindJsxAttributes:
		return node.AsJsxAttributes()
	case ast.KindJsxSpreadAttribute:
		return node.AsJsxSpreadAttribute()
	case ast.KindJsxExpression:
		return node.AsJsxExpression()
	case ast.KindJsxNamespacedName:
		return node.AsJsxNamespacedName()

	// Source file
	case ast.KindSourceFile:
		return node.AsSourceFile()

	// JSDoc
	case ast.KindJSDoc:
		return node.AsJSDoc()
	case ast.KindJSDocTypeExpression:
		return node.AsJSDocTypeExpression()
	case ast.KindJSDocNameReference:
		return node.AsJSDocNameReference()
	case ast.KindJSDocAllType:
		return node.AsJSDocAllType()
	case ast.KindJSDocNullableType:
		return node.AsJSDocNullableType()
	case ast.KindJSDocNonNullableType:
		return node.AsJSDocNonNullableType()
	case ast.KindJSDocOptionalType:
		return node.AsJSDocOptionalType()
	case ast.KindJSDocVariadicType:
		return node.AsJSDocVariadicType()
	case ast.KindJSDocText:
		return node.AsJSDocText()
	case ast.KindJSDocTypeLiteral:
		return node.AsJSDocTypeLiteral()
	case ast.KindJSDocSignature:
		return node.AsJSDocSignature()
	case ast.KindJSDocLink:
		return node.AsJSDocLink()
	case ast.KindJSDocLinkCode:
		return node.AsJSDocLinkCode()
	case ast.KindJSDocLinkPlain:
		return node.AsJSDocLinkPlain()
	case ast.KindJSDocTag:
		return node.AsJSDocTagBase()
	case ast.KindJSDocAugmentsTag:
		return node.AsJSDocAugmentsTag()
	case ast.KindJSDocImplementsTag:
		return node.AsJSDocImplementsTag()
	case ast.KindJSDocDeprecatedTag:
		return node.AsJSDocDeprecatedTag()
	case ast.KindJSDocPublicTag:
		return node.AsJSDocPublicTag()
	case ast.KindJSDocPrivateTag:
		return node.AsJSDocPrivateTag()
	case ast.KindJSDocProtectedTag:
		return node.AsJSDocProtectedTag()
	case ast.KindJSDocReadonlyTag:
		return node.AsJSDocReadonlyTag()
	case ast.KindJSDocOverrideTag:
		return node.AsJSDocOverrideTag()
	case ast.KindJSDocCallbackTag:
		return node.AsJSDocCallbackTag()
	case ast.KindJSDocOverloadTag:
		return node.AsJSDocOverloadTag()
	case ast.KindJSDocParameterTag, ast.KindJSDocPropertyTag:
		return node.AsJSDocParameterOrPropertyTag()
	case ast.KindJSDocReturnTag:
		return node.AsJSDocReturnTag()
	case ast.KindJSDocThisTag:
		return node.AsJSDocThisTag()
	case ast.KindJSDocTypeTag:
		return node.AsJSDocTypeTag()
	case ast.KindJSDocTemplateTag:
		return node.AsJSDocTemplateTag()
	case ast.KindJSDocTypedefTag:
		return node.AsJSDocTypedefTag()
	case ast.KindJSDocSeeTag:
		return node.AsJSDocSeeTag()
	case ast.KindJSDocSatisfiesTag:
		return node.AsJSDocSatisfiesTag()
	case ast.KindJSDocImportTag:
		return node.AsJSDocImportTag()
	default:
		// Token nodes, keywords, and other simple nodes without dedicated structs.
		// Fall back to ForEachChild to discover children.
		return nil
	}
}
