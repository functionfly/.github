import { motion } from "framer-motion";
import { Users, Heart, Shield, Award, Coffee, Code, Lightbulb } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

// Team member data
const teamMembers = [
  {
    name: "Alex Chen",
    role: "Founder & CEO",
    bio: "Former AWS engineer who witnessed too many serverless failures. Built FunctionFly after losing $50K in a single outage.",
    values: ["Reliability First", "Developer Experience", "Transparency"],
    expertise: "Distributed Systems"
  },
  {
    name: "Sarah Rodriguez",
    role: "CTO",
    bio: "Ex-Netflix SRE specializing in chaos engineering. Joined after experiencing the Great Netflix Outage of 2021.",
    values: ["Resilience", "Innovation", "Quality"],
    expertise: "Site Reliability Engineering"
  },
  {
    name: "Marcus Johnson",
    role: "Head of Product",
    bio: "Product leader from Stripe and Twilio. Saw developers struggle with serverless complexity and wanted to fix it.",
    values: ["User-Centric", "Simplicity", "Empowerment"],
    expertise: "Product Strategy"
  },
  {
    name: "Dr. Emily Zhang",
    role: "VP of Engineering",
    bio: "PhD in Computer Science from MIT. Specializes in fault-tolerant distributed systems and loves making complex things simple.",
    values: ["Excellence", "Learning", "Collaboration"],
    expertise: "Systems Architecture"
  }
];

const companyValues = [
  {
    icon: Shield,
    title: "Reliability Above All",
    description: "We believe uptime isn't a feature—it's a fundamental right. Every decision we make prioritizes system stability."
  },
  {
    icon: Heart,
    title: "Developer-First Mindset",
    description: "We build tools we'd want to use ourselves. Your success is our success, and your frustration is our failure."
  },
  {
    icon: Lightbulb,
    title: "Innovation Through Experience",
    description: "Every feature comes from real-world pain points. We learn from outages, not just avoid them."
  },
  {
    icon: Users,
    title: "Transparency & Trust",
    description: "We share our metrics, our mistakes, and our roadmap. Trust is earned through honesty, not marketing."
  }
];

export function TeamSection() {
  return (
    <section className="py-20 border-t border-white/8">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        {/* Section Header */}
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-linear-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center">
              <Users className="w-8 h-8 text-[#6366f1]" />
            </div>
            <h2 className="text-3xl font-bold text-white mb-4">
              Meet the FunctionFly Team
            </h2>
            <p className="text-text-secondary max-w-2xl mx-auto">
              We're not just building monitoring tools—we're solving real problems that have cost companies millions.
              Our story started with frustration, and it's driven by purpose.
            </p>
          </motion.div>
        </div>

        {/* Founder Story */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="mb-20"
        >
          <Card className="border-[#6366f1]/30 bg-[#6366f1]/5 max-w-4xl mx-auto">
            <CardContent className="p-8 lg:p-12">
              <div className="text-center mb-8">
                <h3 className="text-2xl font-bold text-white mb-4">Why We Built FunctionFly</h3>
                <div className="w-24 h-1 bg-[#6366f1] mx-auto rounded-full"></div>
              </div>

              <div className="prose prose-lg prose-invert mx-auto max-w-3xl">
                <p className="text-text-secondary leading-relaxed mb-6">
                  In 2023, I was leading engineering at a fast-growing fintech startup. We had bet big on serverless architecture—Lambda functions, API Gateway, DynamoDB. It seemed like the future: scalable, cost-effective, no servers to manage.
                </p>

                <p className="text-text-secondary leading-relaxed mb-6">
                  Then came the outage. A single misconfigured timeout in one function cascaded through our entire system. What should have been a 5-minute fix turned into a 3-hour nightmare that cost us $50,000 in lost revenue and damaged customer trust.
                </p>

                <p className="text-text-secondary leading-relaxed mb-6">
                  The worst part? Our monitoring tools told us everything was "fine" right up until it wasn't. We had dashboards showing 99.9% uptime, but no insight into what was actually happening with our functions. No early warnings. No automatic recovery. Just expensive lessons learned.
                </p>

                <p className="text-text-secondary leading-relaxed mb-8">
                  That night, I decided to build something better. FunctionFly isn't just another monitoring tool—it's the system we wish we'd had. It's built by engineers who've been through the fire, for engineers who want to sleep soundly at night.
                </p>

                <div className="text-center">
                  <div className="inline-flex items-center gap-2 text-[#6366f1] font-medium">
                    <Coffee className="w-5 h-5" />
                    <span>— Alex Chen, Founder & CEO</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Company Mission */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
          className="mb-20"
        >
          <div className="text-center mb-12">
            <h3 className="text-2xl font-bold text-white mb-4">Our Mission</h3>
            <p className="text-text-secondary max-w-2xl mx-auto">
              To eliminate serverless outages through intelligent monitoring, automatic recovery,
              and unwavering reliability—so developers can focus on building great products instead of fighting fires.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {companyValues.map((value, index) => (
              <motion.div
                key={value.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <Card className="border-white/8 bg-white/5 hover:bg-white/10 transition-colors h-full">
                  <CardContent className="p-6 text-center">
                    <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-[#6366f1]/20 border border-[#6366f1]/20 flex items-center justify-center">
                      <value.icon className="w-6 h-6 text-[#6366f1]" />
                    </div>
                    <h4 className="text-lg font-semibold text-white mb-3">{value.title}</h4>
                    <p className="text-text-secondary text-sm leading-relaxed">{value.description}</p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Team Members */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          <div className="text-center mb-12">
            <h3 className="text-2xl font-bold text-white mb-4">The Team Behind FunctionFly</h3>
            <p className="text-text-secondary max-w-2xl mx-auto">
              We're a small team of experienced engineers who've built and scaled serverless systems at companies
              like AWS, Netflix, Stripe, and Twilio. We know the pain points because we've lived them.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            {teamMembers.map((member, index) => (
              <motion.div
                key={member.name}
                initial={{ opacity: 0, x: index % 2 === 0 ? -20 : 20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <Card className="border-white/8 bg-white/5 hover:bg-white/10 transition-colors">
                  <CardContent className="p-6">
                    <div className="flex items-start gap-4">
                      <div className="w-16 h-16 rounded-full bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center shrink-0">
                        <span className="text-white font-bold text-lg">
                          {member.name.split(' ').map(n => n[0]).join('')}
                        </span>
                      </div>

                      <div className="flex-1 min-w-0">
                        <h4 className="text-lg font-semibold text-white mb-1">{member.name}</h4>
                        <p className="text-[#6366f1] font-medium text-sm mb-3">{member.role}</p>
                        <p className="text-text-secondary text-sm mb-4 leading-relaxed">{member.bio}</p>

                        <div className="space-y-2">
                          <div className="flex items-center gap-2">
                            <Code className="w-4 h-4 text-text-secondary" />
                            <span className="text-text-secondary text-sm">{member.expertise}</span>
                          </div>

                          <div className="flex flex-wrap gap-1">
                            {member.values.map((value) => (
                              <span
                                key={value}
                                className="px-2 py-1 bg-[#6366f1]/20 text-[#6366f1] text-xs rounded-full"
                              >
                                {value}
                              </span>
                            ))}
                          </div>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Call to Action */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.5 }}
          className="text-center mt-16"
        >
          <Card className="border-[#6366f1]/30 bg-[#6366f1]/5 max-w-2xl mx-auto">
            <CardContent className="p-8">
              <Award className="w-12 h-12 text-[#6366f1] mx-auto mb-4" />
              <h3 className="text-xl font-bold text-white mb-3">
                Join Our Mission
              </h3>
              <p className="text-text-secondary mb-6">
                We're always looking for talented engineers who share our passion for building reliable systems.
                If you've experienced the pain of serverless outages and want to help prevent them, we'd love to hear from you.
              </p>
              <a
                href="mailto:careers@functionfly.com"
                className="inline-flex items-center px-6 py-3 rounded-lg bg-[#6366f1] hover:bg-[#6366f1]/80 text-white font-medium transition-colors"
              >
                View Open Positions
              </a>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </section>
  );
}