--!strict
local DataTypes = require("@lib/data_types")

local SourceTable = {}

export type LpSignatoryFields = {
	asaCommitmentAmount: DataTypes.StringType?,
	individualSubscribernameSignaturepage: DataTypes.StringType?,
	entityAuthorizednameSignaturepage: DataTypes.StringType?,
}

export type LpSignatoryType = {
	typeId: string,
	valueSubFields: LpSignatoryFields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

export type W9Fields = {
	w9PartiSsn1: DataTypes.StringType?,
	w9PartiEin1: DataTypes.StringType?,
	w9Line2: DataTypes.StringType?,
}

export type W9Type = {
	typeId: string,
	valueSubFields: W9Fields?,
	subFieldKeysInOrder: {string},
	label: string?,
}

export type SourceTableFieldsMap = {
	lpSignatory: LpSignatoryType?,
	asaFullnameInvestornameAmlquestionnaire: DataTypes.StringType?,
	asaFullnameInvestornameGeneralinfo1: DataTypes.StringType?,
	luxsentityRegulatedstatusPart2Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	indiInternationalsupplementsPart1Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	entityInternationalsupplementsPart1Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	w9: W9Type?,
}

export type SourceTableSchema = {
	fieldsMap: SourceTableFieldsMap?,
	fieldKeysInOrder: {string},
	label: string,
}

function SourceTable.makeLpSignatoryFields(config: {
	asaCommitmentAmount: DataTypes.StringType?,
	individualSubscribernameSignaturepage: DataTypes.StringType?,
	entityAuthorizednameSignaturepage: DataTypes.StringType?,
}?): LpSignatoryFields
	local c = config or {}
	return {
		asaCommitmentAmount = c.asaCommitmentAmount,
		individualSubscribernameSignaturepage = c.individualSubscribernameSignaturepage,
		entityAuthorizednameSignaturepage = c.entityAuthorizednameSignaturepage,
	}
end

function SourceTable.makeLpSignatoryType(config: {
	typeId: string?,
	valueSubFields: LpSignatoryFields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): LpSignatoryType
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function SourceTable.makeW9Fields(config: {
	w9PartiSsn1: DataTypes.StringType?,
	w9PartiEin1: DataTypes.StringType?,
	w9Line2: DataTypes.StringType?,
}?): W9Fields
	local c = config or {}
	return {
		w9PartiSsn1 = c.w9PartiSsn1,
		w9PartiEin1 = c.w9PartiEin1,
		w9Line2 = c.w9Line2,
	}
end

function SourceTable.makeW9Type(config: {
	typeId: string?,
	valueSubFields: W9Fields?,
	subFieldKeysInOrder: {string}?,
	label: string?,
}?): W9Type
	local c = config or {}
	return {
		typeId = c.typeId or "",
		valueSubFields = c.valueSubFields,
		subFieldKeysInOrder = c.subFieldKeysInOrder or {},
		label = c.label,
	}
end

function SourceTable.makeSourceTableFieldsMap(config: {
	lpSignatory: LpSignatoryType?,
	asaFullnameInvestornameAmlquestionnaire: DataTypes.StringType?,
	asaFullnameInvestornameGeneralinfo1: DataTypes.StringType?,
	luxsentityRegulatedstatusPart2Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	indiInternationalsupplementsPart1Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	entityInternationalsupplementsPart1Duediligencequestionnaire: DataTypes.MultipleCheckboxType?,
	w9: W9Type?,
}?): SourceTableFieldsMap
	local c = config or {}
	return {
		lpSignatory = c.lpSignatory,
		asaFullnameInvestornameAmlquestionnaire = c.asaFullnameInvestornameAmlquestionnaire,
		asaFullnameInvestornameGeneralinfo1 = c.asaFullnameInvestornameGeneralinfo1,
		luxsentityRegulatedstatusPart2Duediligencequestionnaire = c.luxsentityRegulatedstatusPart2Duediligencequestionnaire,
		indiInternationalsupplementsPart1Duediligencequestionnaire = c.indiInternationalsupplementsPart1Duediligencequestionnaire,
		entityInternationalsupplementsPart1Duediligencequestionnaire = c.entityInternationalsupplementsPart1Duediligencequestionnaire,
		w9 = c.w9,
	}
end

function SourceTable.makeSourceTableSchema(config: {
	fieldsMap: SourceTableFieldsMap?,
	fieldKeysInOrder: {string}?,
	label: string?,
}?): SourceTableSchema
	local c = config or {}
	return {
		fieldsMap = c.fieldsMap,
		fieldKeysInOrder = c.fieldKeysInOrder or {},
		label = c.label or "",
	}
end

local function decodeOptionalMultipleCheckboxType(json: any): DataTypes.MultipleCheckboxType?
	if json == nil then return nil end
	return DataTypes.decodeJsonMultipleCheckboxType(json)
end
local function encodeOptionalMultipleCheckboxType(v: DataTypes.MultipleCheckboxType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonMultipleCheckboxType(v)
end
local function decodeOptionalStringType(json: any): DataTypes.StringType?
	if json == nil then return nil end
	return DataTypes.decodeJsonStringType(json)
end
local function encodeOptionalStringType(v: DataTypes.StringType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonStringType(v)
end

function SourceTable.encodeJsonLpSignatoryFields(sf: LpSignatoryFields): {[string]: any}
	local result: {[string]: any} = {}
	result.asaCommitmentAmount = encodeOptionalStringType(sf.asaCommitmentAmount)
	result.individualSubscribernameSignaturepage = encodeOptionalStringType(sf.individualSubscribernameSignaturepage)
	result.entityAuthorizednameSignaturepage = encodeOptionalStringType(sf.entityAuthorizednameSignaturepage)
	return result
end

function SourceTable.decodeJsonLpSignatoryFields(json: {[string]: any}): LpSignatoryFields
	return {
		asaCommitmentAmount = decodeOptionalStringType(json.asaCommitmentAmount),
		individualSubscribernameSignaturepage = decodeOptionalStringType(json.individualSubscribernameSignaturepage),
		entityAuthorizednameSignaturepage = decodeOptionalStringType(json.entityAuthorizednameSignaturepage),
	}
end

function SourceTable.encodeJsonLpSignatoryType(t: LpSignatoryType): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = SourceTable.encodeJsonLpSignatoryFields(t.valueSubFields)
	end
	return result
end

function SourceTable.decodeJsonLpSignatoryType(json: {[string]: any}): LpSignatoryType
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then SourceTable.decodeJsonLpSignatoryFields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

function SourceTable.encodeJsonW9Fields(sf: W9Fields): {[string]: any}
	local result: {[string]: any} = {}
	result.w9PartiSsn1 = encodeOptionalStringType(sf.w9PartiSsn1)
	result.w9PartiEin1 = encodeOptionalStringType(sf.w9PartiEin1)
	result.w9Line2 = encodeOptionalStringType(sf.w9Line2)
	return result
end

function SourceTable.decodeJsonW9Fields(json: {[string]: any}): W9Fields
	return {
		w9PartiSsn1 = decodeOptionalStringType(json.w9PartiSsn1),
		w9PartiEin1 = decodeOptionalStringType(json.w9PartiEin1),
		w9Line2 = decodeOptionalStringType(json.w9Line2),
	}
end

function SourceTable.encodeJsonW9Type(t: W9Type): {[string]: any}
	local result: {[string]: any} = {
		typeId = t.typeId,
		subFieldKeysInOrder = t.subFieldKeysInOrder,
	}
	if t.label ~= nil then
		result.label = t.label
	end
	if t.valueSubFields ~= nil then
		result.valueSubFields = SourceTable.encodeJsonW9Fields(t.valueSubFields)
	end
	return result
end

function SourceTable.decodeJsonW9Type(json: {[string]: any}): W9Type
	return {
		typeId = json.typeId or "",
		valueSubFields = if json.valueSubFields then SourceTable.decodeJsonW9Fields(json.valueSubFields) else nil,
		subFieldKeysInOrder = json.subFieldKeysInOrder or {},
		label = json.label,
	}
end

function SourceTable.encodeJsonSourceTableFieldsMap(fm: SourceTableFieldsMap): {[string]: any}
	local result: {[string]: any} = {}
	if fm.lpSignatory then result.lpSignatory = SourceTable.encodeJsonLpSignatoryType(fm.lpSignatory) end
	result.asaFullnameInvestornameAmlquestionnaire = encodeOptionalStringType(fm.asaFullnameInvestornameAmlquestionnaire)
	result.asaFullnameInvestornameGeneralinfo1 = encodeOptionalStringType(fm.asaFullnameInvestornameGeneralinfo1)
	result.luxsentityRegulatedstatusPart2Duediligencequestionnaire = encodeOptionalMultipleCheckboxType(fm.luxsentityRegulatedstatusPart2Duediligencequestionnaire)
	result.indiInternationalsupplementsPart1Duediligencequestionnaire = encodeOptionalMultipleCheckboxType(fm.indiInternationalsupplementsPart1Duediligencequestionnaire)
	result.entityInternationalsupplementsPart1Duediligencequestionnaire = encodeOptionalMultipleCheckboxType(fm.entityInternationalsupplementsPart1Duediligencequestionnaire)
	if fm.w9 then result.w9 = SourceTable.encodeJsonW9Type(fm.w9) end
	return result
end

function SourceTable.decodeJsonSourceTableFieldsMap(json: {[string]: any}): SourceTableFieldsMap
	return {
		lpSignatory = if json.lpSignatory then SourceTable.decodeJsonLpSignatoryType(json.lpSignatory) else nil,
		asaFullnameInvestornameAmlquestionnaire = decodeOptionalStringType(json.asaFullnameInvestornameAmlquestionnaire),
		asaFullnameInvestornameGeneralinfo1 = decodeOptionalStringType(json.asaFullnameInvestornameGeneralinfo1),
		luxsentityRegulatedstatusPart2Duediligencequestionnaire = decodeOptionalMultipleCheckboxType(json.luxsentityRegulatedstatusPart2Duediligencequestionnaire),
		indiInternationalsupplementsPart1Duediligencequestionnaire = decodeOptionalMultipleCheckboxType(json.indiInternationalsupplementsPart1Duediligencequestionnaire),
		entityInternationalsupplementsPart1Duediligencequestionnaire = decodeOptionalMultipleCheckboxType(json.entityInternationalsupplementsPart1Duediligencequestionnaire),
		w9 = if json.w9 then SourceTable.decodeJsonW9Type(json.w9) else nil,
	}
end

function SourceTable.encodeJsonSourceTableSchema(schema: SourceTableSchema): {[string]: any}
	local result: {[string]: any} = {
		fieldKeysInOrder = schema.fieldKeysInOrder,
		label = schema.label,
	}
	if schema.fieldsMap ~= nil then
		result.fieldsMap = SourceTable.encodeJsonSourceTableFieldsMap(schema.fieldsMap)
	end
	return result
end

function SourceTable.decodeJsonSourceTableSchema(json: {[string]: any}): SourceTableSchema
	return {
		fieldsMap = if json.fieldsMap then SourceTable.decodeJsonSourceTableFieldsMap(json.fieldsMap) else nil,
		fieldKeysInOrder = json.fieldKeysInOrder or {},
		label = json.label or "",
	}
end

return SourceTable
