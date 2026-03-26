--!strict
local DataTypes = require("@lib/data_types")

local TargetTable = {}

export type TargetTableFieldsMap = {
	sfAgreementNullCommitmentC: DataTypes.MoneyType?,
	sfAccountSubscriptionInvestorName: DataTypes.StringType?,
	sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC: DataTypes.RadioGroupType?,
	sfAgreementNullWlcInternationalSupplementsC: DataTypes.MultipleCheckboxType?,
	sfAgreementNullSignerFirstName: DataTypes.StringType?,
	sfAgreementNullSignerMiddleName: DataTypes.StringType?,
	sfAgreementNullSignerLastName: DataTypes.StringType?,
	sfTaxFormW9UsTinTypeC: DataTypes.RadioGroupType?,
}

export type TargetTableSchema = {
	fieldsMap: TargetTableFieldsMap?,
	fieldKeysInOrder: {string},
	label: string,
}

function TargetTable.makeTargetTableFieldsMap(config: {
	sfAgreementNullCommitmentC: DataTypes.MoneyType?,
	sfAccountSubscriptionInvestorName: DataTypes.StringType?,
	sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC: DataTypes.RadioGroupType?,
	sfAgreementNullWlcInternationalSupplementsC: DataTypes.MultipleCheckboxType?,
	sfAgreementNullSignerFirstName: DataTypes.StringType?,
	sfAgreementNullSignerMiddleName: DataTypes.StringType?,
	sfAgreementNullSignerLastName: DataTypes.StringType?,
	sfTaxFormW9UsTinTypeC: DataTypes.RadioGroupType?,
}?): TargetTableFieldsMap
	local c = config or {}
	return {
		sfAgreementNullCommitmentC = c.sfAgreementNullCommitmentC,
		sfAccountSubscriptionInvestorName = c.sfAccountSubscriptionInvestorName,
		sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC = c.sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC,
		sfAgreementNullWlcInternationalSupplementsC = c.sfAgreementNullWlcInternationalSupplementsC,
		sfAgreementNullSignerFirstName = c.sfAgreementNullSignerFirstName,
		sfAgreementNullSignerMiddleName = c.sfAgreementNullSignerMiddleName,
		sfAgreementNullSignerLastName = c.sfAgreementNullSignerLastName,
		sfTaxFormW9UsTinTypeC = c.sfTaxFormW9UsTinTypeC,
	}
end

function TargetTable.makeTargetTableSchema(config: {
	fieldsMap: TargetTableFieldsMap?,
	fieldKeysInOrder: {string}?,
	label: string?,
}?): TargetTableSchema
	local c = config or {}
	return {
		fieldsMap = c.fieldsMap,
		fieldKeysInOrder = c.fieldKeysInOrder or {},
		label = c.label or "",
	}
end

local function decodeOptionalMoneyType(json: any): DataTypes.MoneyType?
	if json == nil then return nil end
	return DataTypes.decodeJsonMoneyType(json)
end
local function encodeOptionalMoneyType(v: DataTypes.MoneyType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonMoneyType(v)
end
local function decodeOptionalMultipleCheckboxType(json: any): DataTypes.MultipleCheckboxType?
	if json == nil then return nil end
	return DataTypes.decodeJsonMultipleCheckboxType(json)
end
local function encodeOptionalMultipleCheckboxType(v: DataTypes.MultipleCheckboxType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonMultipleCheckboxType(v)
end
local function decodeOptionalRadioGroupType(json: any): DataTypes.RadioGroupType?
	if json == nil then return nil end
	return DataTypes.decodeJsonRadioGroupType(json)
end
local function encodeOptionalRadioGroupType(v: DataTypes.RadioGroupType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonRadioGroupType(v)
end
local function decodeOptionalStringType(json: any): DataTypes.StringType?
	if json == nil then return nil end
	return DataTypes.decodeJsonStringType(json)
end
local function encodeOptionalStringType(v: DataTypes.StringType?): {[string]: any}?
	if v == nil then return nil end
	return DataTypes.encodeJsonStringType(v)
end

function TargetTable.encodeJsonTargetTableFieldsMap(ta: TargetTableFieldsMap): {[string]: any}
	local result: {[string]: any} = {
	}
	result.sfAgreementNullCommitmentC = encodeOptionalMoneyType(ta.sfAgreementNullCommitmentC)
	result.sfAccountSubscriptionInvestorName = encodeOptionalStringType(ta.sfAccountSubscriptionInvestorName)
	result.sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC = encodeOptionalRadioGroupType(ta.sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC)
	result.sfAgreementNullWlcInternationalSupplementsC = encodeOptionalMultipleCheckboxType(ta.sfAgreementNullWlcInternationalSupplementsC)
	result.sfAgreementNullSignerFirstName = encodeOptionalStringType(ta.sfAgreementNullSignerFirstName)
	result.sfAgreementNullSignerMiddleName = encodeOptionalStringType(ta.sfAgreementNullSignerMiddleName)
	result.sfAgreementNullSignerLastName = encodeOptionalStringType(ta.sfAgreementNullSignerLastName)
	result.sfTaxFormW9UsTinTypeC = encodeOptionalRadioGroupType(ta.sfTaxFormW9UsTinTypeC)
	return result
end

function TargetTable.decodeJsonTargetTableFieldsMap(json: {[string]: any}): TargetTableFieldsMap
	return {
		sfAgreementNullCommitmentC = decodeOptionalMoneyType(json.sfAgreementNullCommitmentC),
		sfAccountSubscriptionInvestorName = decodeOptionalStringType(json.sfAccountSubscriptionInvestorName),
		sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC = decodeOptionalRadioGroupType(json.sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC),
		sfAgreementNullWlcInternationalSupplementsC = decodeOptionalMultipleCheckboxType(json.sfAgreementNullWlcInternationalSupplementsC),
		sfAgreementNullSignerFirstName = decodeOptionalStringType(json.sfAgreementNullSignerFirstName),
		sfAgreementNullSignerMiddleName = decodeOptionalStringType(json.sfAgreementNullSignerMiddleName),
		sfAgreementNullSignerLastName = decodeOptionalStringType(json.sfAgreementNullSignerLastName),
		sfTaxFormW9UsTinTypeC = decodeOptionalRadioGroupType(json.sfTaxFormW9UsTinTypeC),
	}
end

function TargetTable.encodeJsonTargetTableSchema(ta: TargetTableSchema): {[string]: any}
	local result: {[string]: any} = {
		fieldKeysInOrder = ta.fieldKeysInOrder,
		label = ta.label,
	}
	if ta.fieldsMap ~= nil then
		result.fieldsMap = TargetTable.encodeJsonTargetTableFieldsMap(ta.fieldsMap)
	end
	return result
end

function TargetTable.decodeJsonTargetTableSchema(json: {[string]: any}): TargetTableSchema
	return {
		fieldsMap = if json.fieldsMap then TargetTable.decodeJsonTargetTableFieldsMap(json.fieldsMap) else nil,
		fieldKeysInOrder = json.fieldKeysInOrder or {},
		label = json.label or "",
	}
end

return TargetTable
