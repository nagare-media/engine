/*
Copyright 2022-2025 The nagare media authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"k8s.io/utils/ptr"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func DefaultFunction(f *nbmpv2.Function) error {
	if f.Scheme == nil {
		f.Scheme = &nbmpv2.Scheme{}
	}
	if err := DefaultScheme(f.Scheme); err != nil {
		return err
	}
	if err := DefaultGeneral(&f.General); err != nil {
		return err
	}
	if err := DefaultInput(&f.Input); err != nil {
		return err
	}
	if err := DefaultOutput(&f.Output); err != nil {
		return err
	}
	if f.Processing != nil {
		if err := DefaultProcessing(f.Processing); err != nil {
			return err
		}
	}
	if f.Requirement != nil {
		if err := DefaultRequirement(f.Requirement); err != nil {
			return err
		}
	}
	if f.Step != nil {
		if err := DefaultStep(f.Step); err != nil {
			return err
		}
	}
	if f.ClientAssistant != nil {
		if err := DefaultClientAssistant(f.ClientAssistant); err != nil {
			return err
		}
	}
	if f.Assertion != nil {
		if err := DefaultAssertion(f.Assertion); err != nil {
			return err
		}
	}
	if f.Security != nil {
		if err := DefaultSecurity(f.Security); err != nil {
			return err
		}
	}
	return nil
}

func DefaultMediaProcessingEntityCapabilities(mpec *nbmpv2.MediaProcessingEntityCapabilities) error {
	if mpec.Scheme == nil {
		mpec.Scheme = &nbmpv2.Scheme{}
	}
	if err := DefaultScheme(mpec.Scheme); err != nil {
		return err
	}
	if err := DefaultGeneral(&mpec.General); err != nil {
		return err
	}
	if mpec.Capabilities != nil {
		if err := DefaultCapabilities(mpec.Capabilities); err != nil {
			return err
		}
	}
	if mpec.Reporting != nil {
		if err := DefaultReporting(mpec.Reporting); err != nil {
			return err
		}
	}
	if mpec.Notification != nil {
		if err := DefaultNotification(mpec.Notification); err != nil {
			return err
		}
	}
	return nil
}

func DefaultWorkflow(w *nbmpv2.Workflow) error {
	if w.Scheme == nil {
		w.Scheme = &nbmpv2.Scheme{}
	}
	if err := DefaultScheme(w.Scheme); err != nil {
		return err
	}
	if err := DefaultGeneral(&w.General); err != nil {
		return err
	}
	if w.Repository != nil {
		if err := DefaultRepository(w.Repository); err != nil {
			return err
		}
	}
	if err := DefaultInput(&w.Input); err != nil {
		return err
	}
	if err := DefaultOutput(&w.Output); err != nil {
		return err
	}
	if err := DefaultProcessing(&w.Processing); err != nil {
		return err
	}
	if err := DefaultRequirement(&w.Requirement); err != nil {
		return err
	}
	if w.Step != nil {
		if err := DefaultStep(w.Step); err != nil {
			return err
		}
	}
	if w.ClientAssistant != nil {
		if err := DefaultClientAssistant(w.ClientAssistant); err != nil {
			return err
		}
	}
	if w.Failover != nil {
		if err := DefaultFailover(w.Failover); err != nil {
			return err
		}
	}
	if w.Assertion != nil {
		if err := DefaultAssertion(w.Assertion); err != nil {
			return err
		}
	}
	if w.Reporting != nil {
		if err := DefaultReporting(w.Reporting); err != nil {
			return err
		}
	}
	if w.Notification != nil {
		if err := DefaultNotification(w.Notification); err != nil {
			return err
		}
	}
	// always reset Acknowledge for response
	w.Acknowledge = &nbmpv2.Acknowledge{}
	if err := DefaultAcknowledge(w.Acknowledge); err != nil {
		return err
	}
	if w.Security != nil {
		if err := DefaultSecurity(w.Security); err != nil {
			return err
		}
	}
	if w.Scale != nil {
		if err := DefaultScale(w.Scale); err != nil {
			return err
		}
	}
	if w.Schedule != nil {
		if err := DefaultSchedule(w.Schedule); err != nil {
			return err
		}
	}
	return nil
}

func DefaultTask(t *nbmpv2.Task) error {
	if t.Scheme == nil {
		t.Scheme = &nbmpv2.Scheme{}
	}
	if err := DefaultScheme(t.Scheme); err != nil {
		return err
	}
	if err := DefaultGeneral(&t.General); err != nil {
		return err
	}
	if err := DefaultInput(&t.Input); err != nil {
		return err
	}
	if err := DefaultOutput(&t.Output); err != nil {
		return err
	}
	if err := DefaultProcessing(&t.Processing); err != nil {
		return err
	}
	if err := DefaultRequirement(&t.Requirement); err != nil {
		return err
	}
	if t.Step != nil {
		if err := DefaultStep(t.Step); err != nil {
			return err
		}
	}
	if t.ClientAssistant != nil {
		if err := DefaultClientAssistant(t.ClientAssistant); err != nil {
			return err
		}
	}
	if t.Failover != nil {
		if err := DefaultFailover(t.Failover); err != nil {
			return err
		}
	}
	if t.Assertion != nil {
		if err := DefaultAssertion(t.Assertion); err != nil {
			return err
		}
	}
	if t.Reporting != nil {
		if err := DefaultReporting(t.Reporting); err != nil {
			return err
		}
	}
	if t.Notification != nil {
		if err := DefaultNotification(t.Notification); err != nil {
			return err
		}
	}
	// always reset Acknowledge for response
	t.Acknowledge = &nbmpv2.Acknowledge{}
	if err := DefaultAcknowledge(t.Acknowledge); err != nil {
		return err
	}
	if t.Security != nil {
		if err := DefaultSecurity(t.Security); err != nil {
			return err
		}
	}
	if t.Scale != nil {
		if err := DefaultScale(t.Scale); err != nil {
			return err
		}
	}
	if t.Schedule != nil {
		if err := DefaultSchedule(t.Schedule); err != nil {
			return err
		}
	}
	return nil
}

func DefaultScheme(s *nbmpv2.Scheme) error {
	if s.URI == "" {
		s.URI = nbmpv2.SchemaURI
	}
	return nil
}

func DefaultGeneral(g *nbmpv2.General) error {
	if g.IsGroup == nil {
		g.IsGroup = ptr.To(false)
	}
	if g.Nonessential == nil {
		g.Nonessential = ptr.To(false)
	}
	for i := range g.TaskGroup {
		if err := DefaultTaskGroupItem(&g.TaskGroup[i]); err != nil {
			return err
		}
	}
	return nil
}

func DefaultTaskGroupItem(tgi *nbmpv2.TaskGroupItem) error {
	if tgi.GroupType == nil {
		tgi.GroupType = ptr.To(nbmpv2.DistanceGroupType)
	}
	if tgi.GroupMode == nil {
		tgi.GroupMode = ptr.To(nbmpv2.SynchronousGroupMode)
	}
	if tgi.NetZero == nil {
		tgi.NetZero = ptr.To(false)
	}
	return nil
}

func DefaultInput(i *nbmpv2.Input) error {
	for j := range i.MediaParameters {
		if err := DefaultMediaParameter(&i.MediaParameters[j]); err != nil {
			return err
		}
	}
	for j := range i.MetadataParameters {
		if err := DefaultMetadataParameter(&i.MetadataParameters[j]); err != nil {
			return err
		}
	}
	return nil
}

func DefaultOutput(o *nbmpv2.Output) error {
	for j := range o.MediaParameters {
		if err := DefaultMediaParameter(&o.MediaParameters[j]); err != nil {
			return err
		}
	}
	for j := range o.MetadataParameters {
		if err := DefaultMetadataParameter(&o.MetadataParameters[j]); err != nil {
			return err
		}
	}
	return nil
}

func DefaultMediaParameter(mp *nbmpv2.MediaParameter) error {
	if mp.Mode == nil {
		mp.Mode = ptr.To(nbmpv2.PushMediaAccessMode)
	}
	return nil
}

func DefaultMetadataParameter(mp *nbmpv2.MetadataParameter) error {
	if mp.Mode == nil {
		mp.Mode = ptr.To(nbmpv2.PushMediaAccessMode)
	}
	return nil
}

func DefaultProcessing(p *nbmpv2.Processing) error {
	for i := range p.Image {
		if err := DefaultProcessingImage(&p.Image[i]); err != nil {
			return err
		}
	}

	for i := range p.ConnectionMap {
		if err := DefaultConnectionMapping(&p.ConnectionMap[i]); err != nil {
			return err
		}
	}

	for i := range p.FunctionRestrictions {
		if err := DefaultFunctionRestriction(&p.FunctionRestrictions[i]); err != nil {
			return err
		}
	}

	return nil
}

func DefaultProcessingImage(pi *nbmpv2.ProcessingImage) error {
	if pi.IsDynamic == nil {
		pi.IsDynamic = ptr.To(false)
	}
	return nil
}

func DefaultConnectionMapping(cm *nbmpv2.ConnectionMapping) error {
	if cm.CoLocated == nil {
		cm.CoLocated = ptr.To(false)
	}
	if cm.Breakable == nil {
		cm.Breakable = ptr.To(false)
	}
	if cm.Flowcontrol != nil {
		if err := DefaultFlowcontrolRequirement(cm.Flowcontrol); err != nil {
			return err
		}
	}
	if err := DefaultConnectionMappingPort(&cm.From); err != nil {
		return err
	}
	if err := DefaultConnectionMappingPort(&cm.To); err != nil {
		return err
	}
	return nil
}

func DefaultConnectionMappingPort(p *nbmpv2.ConnectionMappingPort) error {
	if p.InputRestrictions != nil {
		if err := DefaultInput(p.InputRestrictions); err != nil {
			return err
		}
	}
	if p.OutputRestrictions != nil {
		if err := DefaultOutput(p.OutputRestrictions); err != nil {
			return err
		}
	}
	return nil
}

func DefaultFunctionRestriction(fr *nbmpv2.FunctionRestriction) error {
	if fr.General != nil {
		if err := DefaultGeneral(fr.General); err != nil {
			return err
		}
	}
	if fr.Processing != nil {
		if err := DefaultProcessing(fr.Processing); err != nil {
			return err
		}
	}
	if fr.Requirements != nil {
		if err := DefaultRequirement(fr.Requirements); err != nil {
			return err
		}
	}
	if fr.ClientAssistant != nil {
		if err := DefaultClientAssistant(fr.ClientAssistant); err != nil {
			return err
		}
	}
	if fr.Failover != nil {
		if err := DefaultFailover(fr.Failover); err != nil {
			return err
		}
	}
	if fr.Reporting != nil {
		if err := DefaultReporting(fr.Reporting); err != nil {
			return err
		}
	}
	if fr.Notification != nil {
		if err := DefaultNotification(fr.Notification); err != nil {
			return err
		}
	}
	if fr.Step != nil {
		if err := DefaultStep(fr.Step); err != nil {
			return err
		}
	}
	if fr.Security != nil {
		if err := DefaultSecurity(fr.Security); err != nil {
			return err
		}
	}
	return nil
}

func DefaultRequirement(r *nbmpv2.Requirement) error {
	if r.Flowcontrol != nil {
		if err := DefaultFlowcontrolRequirement(r.Flowcontrol); err != nil {
			return err
		}
	}
	if r.Security != nil {
		if err := DefaultSecurityRequirement(r.Security); err != nil {
			return err
		}
	}
	if r.WorkflowTask != nil {
		if err := DefaultWorkflowTaskRequirement(r.WorkflowTask); err != nil {
			return err
		}
	}
	return nil
}

func DefaultFlowcontrolRequirement(fcr *nbmpv2.FlowcontrolRequirement) error {
	// TODO: set
	// if fcr.AveragingWindow == nil {
	// 	fcr.AveragingWindow = ptr.To(uint64(1000)) // = 1s
	// }
	return nil
}

func DefaultSecurityRequirement(sr *nbmpv2.SecurityRequirement) error {
	if sr.TLS == nil {
		sr.TLS = ptr.To(false)
	}
	if sr.IPsec == nil {
		sr.IPsec = ptr.To(false)
	}
	if sr.CENC == nil {
		sr.CENC = ptr.To(false)
	}
	return nil
}

func DefaultWorkflowTaskRequirement(wtr *nbmpv2.WorkflowTaskRequirement) error {
	if wtr.FunctionFusible == nil {
		wtr.FunctionFusible = ptr.To(false)
	}
	if wtr.FunctionEnhancable == nil {
		wtr.FunctionEnhancable = ptr.To(false)
	}
	if wtr.ExecutionMode == nil {
		wtr.ExecutionMode = ptr.To(nbmpv2.StreamingExecutionMode)
	}
	if wtr.SplitEfficiency != nil {
		if err := DefaultTaskSplitEfficiency(wtr.SplitEfficiency); err != nil {
			return err
		}
	}
	return nil
}

func DefaultTaskSplitEfficiency(tse *nbmpv2.TaskSplitEfficiency) error {
	if tse.SplitNorm == nil {
		tse.SplitNorm = ptr.To(nbmpv2.PnormTaskSplitEfficiencyNorm)
	}
	if tse.SplitEquation == nil {
		tse.SplitEquation = ptr.To("2")
	}
	return nil
}

func DefaultClientAssistant(ca *nbmpv2.ClientAssistant) error {
	// ca.ClientAssistanceFlag defaults to false
	return nil
}

func DefaultFailover(fo *nbmpv2.Failover) error {
	if fo.FailoverMode == "" {
		fo.FailoverMode = nbmpv2.ExitFailoverMode
	}
	return nil
}

func DefaultReporting(r *nbmpv2.Reporting) error {
	if r.DeliveryMethod == "" {
		r.DeliveryMethod = nbmpv2.HTTP_POSTDeliveryMethod
	}
	return nil
}

func DefaultNotification(n *nbmpv2.Notification) error {
	if n.NotificationInterval == nil {
		n.NotificationInterval = ptr.To(uint64(0))
	}
	return nil
}

func DefaultAssertion(a *nbmpv2.Assertion) error {
	if a.SupportVerification == nil {
		a.SupportVerification = ptr.To(false)
	}
	return nil
}

func DefaultAcknowledge(a *nbmpv2.Acknowledge) error {
	if a.Unsupported == nil {
		a.Unsupported = make([]string, 0)
	}
	if a.Failed == nil {
		a.Failed = make([]string, 0)
	}
	if a.Partial == nil {
		a.Partial = make([]string, 0)
	}
	return nil
}

func DefaultRepository(r *nbmpv2.Repository) error {
	if r.Mode == nil {
		r.Mode = ptr.To(nbmpv2.AvailableRepositoryMode)
	}
	return nil
}

func DefaultSecurity(s *nbmpv2.Security) error {
	if s.Scope == nil {
		s.Scope = ptr.To(nbmpv2.DataSecurityScope)
	}
	if s.AuthTokenRotation == nil {
		s.AuthTokenRotation = ptr.To(false)
	}
	return nil
}

func DefaultStep(s *nbmpv2.Step) error {
	if s.StepMode == nil {
		s.StepMode = ptr.To(nbmpv2.StreamStepMode)
	}
	if s.VariableDuration == nil {
		s.VariableDuration = ptr.To(false)
	}
	if s.SegmentLocation == nil {
		s.SegmentLocation = ptr.To(false)
	}
	if s.SegmentSequence == nil {
		s.SegmentSequence = ptr.To(false)
	}
	return nil
}

func DefaultCapabilities(c *nbmpv2.Capabilities) error {
	if c.Repository != nil {
		if err := DefaultRepository(c.Repository); err != nil {
			return nil
		}
	}
	for i := range c.Functions {
		if err := DefaultFunction(&c.Functions[i]); err != nil {
			return err
		}
	}
	if c.PersistencyCapabilities == nil {
		c.PersistencyCapabilities = ptr.To(true)
	}
	if c.SecurePersistency == nil {
		c.SecurePersistency = ptr.To(true)
	}
	return nil
}

func DefaultScale(s *nbmpv2.Scale) error {
	if s.ScalingFactor == nil {
		s.ScalingFactor = ptr.To(uint64(1))
	}
	if s.Status == "" {
		s.Status = nbmpv2.FailedScalingStatus
	}
	return nil
}

func DefaultSchedule(s *nbmpv2.Schedule) error {
	if s.ScheduleType == nil {
		s.ScheduleType = ptr.To(nbmpv2.DurationScheduleType)
	}
	for i := range s.ScheduleTable {
		if err := DefaultScheduleTableItem(&s.ScheduleTable[i]); err != nil {
			return err
		}
	}

	if s.NumberOfSegments == nil {
		s.NumberOfSegments = ptr.To(uint64(1))
	}
	if s.Loop == nil {
		s.Loop = ptr.To(false)
	}
	if s.Status == "" {
		s.Status = nbmpv2.FailedScheduleStatus
	}
	return nil
}

func DefaultScheduleTableItem(sti *nbmpv2.ScheduleTableItem) error {
	if sti.Duration == nil {
		sti.Duration = ptr.To(uint64(1))
	}
	if sti.Timescale == nil {
		sti.Timescale = ptr.To(uint64(1))
	}
	return nil
}
