

const Footer = () => {
  return (
    <footer className="bg-[#0F172A] text-white py-6 mt-12 h-56">
      <div className="container mx-auto px-4 text-center">
        <p>&copy; {new Date().getFullYear()} JobConnect. All rights reserved.</p>
      </div>
    </footer>
  );
};

export default Footer;