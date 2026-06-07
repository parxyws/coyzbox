package helper

import (
	"fmt"
	"strings"

	"github.com/parxyws/cozybox/internal/domain"
)

const (
	IdPrefixQuotation     = "QUO"
	IdPrefixInvoice       = "INV"
	IdPrefixReceipt       = "REC"
	IdPrefixPurchaseOrder = "PO"
	IdPrefixSalesOrder    = "SO"
	IdPrefixDebitNote     = "DN"
)

func getLastId() string {
	return ""
}

func IdentificationGenerator(uniqueId string, documentType domain.DocumentType) string {
	var documentTypeId string

	switch documentType {
	case domain.DocumentTypeQuotation:
		documentTypeId = IdPrefixQuotation
	case domain.DocumentTypeInvoice:
		documentTypeId = IdPrefixInvoice
	case domain.DocumentTypeReceipt:
		documentTypeId = IdPrefixReceipt
	case domain.DocumentTypePurchaseOrder:
		documentTypeId = IdPrefixPurchaseOrder
	case domain.DocumentTypeSalesOrder:
		documentTypeId = IdPrefixSalesOrder
	case domain.DocumentTypeDebitNote:
		documentTypeId = IdPrefixDebitNote
	}

	lastId := strings.Split(getLastId(), "/")
	generatedId := fmt.Sprintf("%s-%s-%s", documentTypeId, uniqueId, lastId)
	return generatedId
}
