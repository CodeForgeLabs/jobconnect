
// bg-surface-container-low 
const Footer = () => {
  return (
    <footer className="bg-[#0F172A] text-white h-56 py-16 w-full flex flex-col md:flex-row justify-between items-center px-12 gap-8">
       <div className="flex flex-col gap-4">
          <span className="font-headline text-xl font-bold text-secondary">
            JobConnect
          </span>
          <p className="font-body text-xs text-on-surface-variant">
            © 2024 JobConnect. Architecting the future of work.
          </p>
        </div>
        <div className="flex gap-12">
          <div className="flex flex-col gap-3">
            <span className="font-label text-xs font-bold text-on-surface-variant uppercase tracking-widest">
              Company
            </span>
            <a
              className="font-body text-xs text-on-surface-variant hover:text-primary transition-all hover:underline underline-offset-4"
              href="#"
            >
              About
            </a>
            <a
              className="font-body text-xs text-on-surface-variant hover:text-primary transition-all hover:underline underline-offset-4"
              href="#"
            >
              Privacy
            </a>
            <a
              className="font-body text-xs text-on-surface-variant hover:text-primary transition-all hover:underline underline-offset-4"
              href="#"
            >
              Terms
            </a>
          </div>
          <div className="flex flex-col gap-3">
            <span className="font-label text-xs font-bold text-on-surface-variant uppercase tracking-widest">
              Support
            </span>
            <a
              className="font-body text-xs text-on-surface-variant hover:text-primary transition-all hover:underline underline-offset-4"
              href="#"
            >
              Help Center
            </a>
            <a
              className="font-body text-xs text-on-surface-variant hover:text-primary transition-all hover:underline underline-offset-4"
              href="#"
            >
              Contact
            </a>
          </div>
        </div>
    </footer>
  );
};

export default Footer;