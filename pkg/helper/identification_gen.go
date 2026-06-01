package helper

import (
	"fmt"
	"strings"

	"github.com/parxyws/cozybox/internal/models"
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

func IdentificationGenerator(uniqueId string, documentType models.DocumentType) string {
	var documentTypeId string
	lastId := strings.Split(getLastId(), "/")

	switch documentType {
	case models.DocumentTypeQuotation:
		documentTypeId = IdPrefixQuotation
	case models.DocumentTypeInvoice:
		documentTypeId = IdPrefixInvoice
	case models.DocumentTypeReceipt:
		documentTypeId = IdPrefixReceipt
	case models.DocumentTypePurchaseOrder:
		documentTypeId = IdPrefixPurchaseOrder
	case models.DocumentTypeSalesOrder:
		documentTypeId = IdPrefixSalesOrder
	case models.DocumentTypeDebitNote:
		documentTypeId = IdPrefixDebitNote
	}

	generatedId := fmt.Sprintf("%s-%s-%s", documentTypeId, uniqueId, lastId)

	return generatedId
}
