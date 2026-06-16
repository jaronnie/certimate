import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconWorldSearch } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizMonitorDomainNodeConfigForm from "../forms/BizMonitorDomainNodeConfigForm";

export const BizMonitorDomainNodeRegistry: NodeRegistry = {
  type: NodeType.BizMonitorDomain,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.monitor_domain.label"),

    icon: IconWorldSearch,
    iconColor: "#fff",
    iconBgColor: "#13a8a8",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizMonitorDomainNodeConfigForm.getSchema({}).safeParse(value);
        if (!res.success) {
          return {
            message: res.error.message,
            level: FeedbackLevel.Error,
          };
        }
      },
    },

    render: () => {
      const { t } = getI18n();

      return (
        <BaseNode
          description={
            <Field name="config.domain">
              {({ field: { value } }) => <>{value ? value : t("workflow.detail.design.editor.placeholder")}</>}
            </Field>
          }
        />
      );
    },
  },

  onAdd: () => {
    return newNode(NodeType.BizMonitorDomain, { i18n: getI18n() });
  },
};
